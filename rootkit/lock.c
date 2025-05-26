#include "lock.h"

#include <linux/string.h>

static char is_locked = LOCKED;

int lock_status(void)
{
    return is_locked;
}

int try_unlock(char *password)
{
    if (password && strcmp("oui", password) == 0)
    {
        is_locked = UNLOCKED;
    }
    return is_locked;
}

void lock(void)
{
  is_locked = LOCKED;
}
