#if !defined(__MREF_H__)
#define __MREF_H__

#include "comm.h"

int mref_init(void);

void *mref_alloc(size_t size, void *pool, mem_alloc_cb_t alloc, mem_dealloc_cb_t dealloc);
void mref_dealloc(void *pool, void *addr);

int mref_check(void *addr);

int mref_inc(void *addr);
int mref_dec(void *addr);

#endif /*__MREF_H__*/
