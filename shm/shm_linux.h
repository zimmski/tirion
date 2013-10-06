#ifndef shm_linux_h_INCLUDED
#define shm_linux_h_INCLUDED

float* shmAttach(long shm_id);
long shmClose(long shm_id);
void shmCopy(float* from, float* to, long count);
long shmDetach(float *addr);
float shmGet(float* addr, long i);
long shmOpen(char* filename, char create, long count);
float shmSet(float *addr, long i, float v);

float shmAdd(float* addr, long i, float v);
float shmDec(float* addr, long i);
float shmInc(float* addr, long i);
float shmSub(float* addr, long i, float v);

#endif
