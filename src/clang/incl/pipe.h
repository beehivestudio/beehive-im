#if !defined(__PIPE_H__)
#define __PIPE_H__

typedef struct
{
    int fd[2];      /* 文件描述符(0:读 1:写) */
} pipe_t;

int pipe_creat(pipe_t *p);

#define pipe_read(p, buf, count) read((p)->fd[0], buf, count)
#define pipe_write(p, buf, count) write((p)->fd[1], buf, count)

#endif /*__PIPE_H__*/
