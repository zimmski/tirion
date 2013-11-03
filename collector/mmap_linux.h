#ifndef mmap_linux_h_INCLUDED
#define mmap_linux_h_INCLUDED

float *mmapOpen(const char *filename, char create, long count);
int mmapClose(float *addr, const char *filename, char create, long count);
void mmapCopy(float* from, float* to, long count);

float mmapAdd(float* addr, long i, float v);
float mmapDec(float* addr, long i);
float mmapInc(float* addr, long i);
float mmapSub(float* addr, long i, float v);

#endif
