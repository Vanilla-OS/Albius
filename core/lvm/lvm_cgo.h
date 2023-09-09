#ifndef LVM_CGO_H
#define LVM_CGO_H

#include <lvm2cmd.h>
#include <stdlib.h>
#include <string.h>

typedef struct Lvm_log_
{
    struct Lvm_log_ *next_;
    void *entry_;
} Lvm_log;

Lvm_log *lvm_log_new();
void lvm_log_push(Lvm_log **log, void *restrict entry);
void *lvm_log_remove(Lvm_log **log);
int lvm_log_empty(Lvm_log *log);

Lvm_log *logger();
void set_logger(Lvm_log *log);

void init_logger();

void lvm_log_capture_fn(int level, const char *file, int line,
                        int dm_errno, const char *format);

#endif