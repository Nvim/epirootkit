#ifndef LOCK_H
#define LOCK_H

// similar to hide, keep a global boolean to know wether we're locked,
// and provide access with a function

#define LOCKED 1
#define UNLOCKED 0

// 1 = locked, 0 = unlocked
int lock_status(void);

// returns is_locked
int try_unlock(char *password);

void lock(void);

#endif
