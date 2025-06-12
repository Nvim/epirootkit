#include "lock.h"

#include <linux/module.h>
#include <linux/string.h>

#include "b64.h"

static char is_locked = LOCKED;

static const char *password = "dGhpc2lzdmVyeXNlY3VyZQ==";

int lock_status(void)
{
    return is_locked;
}

int try_unlock(char *input)
{
    char *decoded = b64_encode(input, strlen(input));
    if (input && strcmp(password, decoded) == 0)
    {
        is_locked = UNLOCKED;
    }
    else
    {
        pr_warn("lock: invalid password: %s doesn't match %s\n", decoded,
                password);
    }
    return is_locked;
}

void lock(void)
{
    is_locked = LOCKED;
}
