#ifndef EXEC_H

#define EXEC_H

char *get_last_outfile(void);
char *get_last_errfile(void);

// Run command synchronously
int exec_sync(const char *cmd_str, int *ret);

#endif // !EXEC_H
