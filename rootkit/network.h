#ifndef NETWORK_H
#define NETWORK_H

int network_init(const char *ip, int port);
int network_loop(void *data);
void network_exit(void);

#endif
