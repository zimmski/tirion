#ifndef shm_linux_h_INCLUDED
#define shm_linux_h_INCLUDED

float* shmAttach(int shm_id);
int shmClose(int shm_id);
void shmCopy(float* from, float* to, int count);
int shmDetach(float *addr);
float shmGet(float* addr, int i);
int shmOpen(char* filename, int create, int count);
float shmSet(float *addr, int i, float v);

float shmAdd(float* addr, int i, float v);
float shmDec(float* addr, int i);
float shmInc(float* addr, int i);
float shmSub(float* addr, int i, float v);

#endif
