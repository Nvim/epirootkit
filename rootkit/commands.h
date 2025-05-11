#ifndef COMMANDS_H
#define COMMANDS_H

#define MAX_ARGS_LEN 1024

enum command_type
{
    CMD_EXEC_SYNC,
    CMD_EXEC_ASYNC,
    CMD_HIDE,
    CMD_UPLOAD,
    CMD_DOWNLOAD,
    CMD_UNKNOWN
};

typedef void (*cmd_callback)(char *args, char *status_buf);

struct command
{
    enum command_type type;
    char args[MAX_ARGS_LEN];
    cmd_callback callback;
    // int status; // stores callback's return value
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
