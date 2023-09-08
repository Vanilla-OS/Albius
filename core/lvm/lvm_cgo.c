#include "lvm_cgo.h"

char *log_output_;

void lvm_log_capture_fn(int level, const char *file, int line,
						int dm_errno, const char *format)
{
	if (level != 4)
		return;

	size_t log_len = strlen(format)+1;
	log_output_ = (char *)malloc(log_len * sizeof(char));
	memcpy(log_output_, format, log_len);

	return;
}

char **log_output() {
	return &log_output_;
}