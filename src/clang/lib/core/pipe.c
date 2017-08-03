#include "sck.h"
#include "comm.h"
#include "pipe.h"

int pipe_creat(pipe_t *p)
{
    if (pipe(p->fd)) {
        return -1;
    }

    fd_set_nonblocking(p->fd[0]);
    fd_set_nonblocking(p->fd[1]);

    return 0;
}
