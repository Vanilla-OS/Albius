#include "lvm_cgo.h"
#include <stdio.h>

Lvm_log *lvm_log_new()
{
    Lvm_log *log = malloc(sizeof(Lvm_log));
    log->next_   = NULL;
    log->entry_  = NULL;
    return log;
}

void lvm_log_push(Lvm_log **log, void *restrict entry)
{
    Lvm_log *l = *log;
    while (l->entry_ != NULL)
        l = l->next_;

    l->next_  = lvm_log_new();
    l->entry_ = entry;
}

void *restrict lvm_log_remove(Lvm_log **log)
{
    void *entry       = (*log)->entry_;
    Lvm_log *next_log = (*log)->next_;
    free(*log);
    *log = next_log;

    return entry;
}

int lvm_log_empty(Lvm_log *log)
{
    return log->entry_ == NULL;
}

Lvm_log *logger_;

Lvm_log *logger()
{
    return logger_;
}

void set_logger(Lvm_log *log)
{
    logger_ = log;
}

void init_logger()
{
    if (logger_ == NULL)
        logger_ = lvm_log_new();
}

void lvm_log_capture_fn(int level, const char *file, int line,
                        int dm_errno, const char *format)
{
    if (level != 4)
        return;

    size_t log_len = strlen(format);

    char *out = (char *)malloc(log_len * sizeof(char));
    memcpy(out, format, log_len + 1);

    lvm_log_push(&logger_, out);

    return;
}