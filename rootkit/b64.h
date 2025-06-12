#ifndef B64_h

#define B64_h

#include <linux/module.h>

char *b64_encode(const unsigned char *in, size_t len);

#endif // !B64_h
