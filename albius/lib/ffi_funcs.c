#include "ffi_funcs.h"

void _ffi_println(char *s) {
    printf("%s\n", s);
}

void add_path_to_partitions(partition *parts, int count, char *path) {
    for(int i = 0; i < count; i++) {
        char *path_prefix;
        char num_str[3];

        if (atoi(&path[strlen(path)-1]) == 0)
            path_prefix = strdup(path);
        else {
            path_prefix = (char *)malloc((strlen(path) + 1) * sizeof(char));
            strcpy(path_prefix, path);
            strcat(path_prefix, "p");
        }

        sprintf(num_str, "%d", parts[i]._number);

        parts[i]._path = (char *)malloc((strlen(path_prefix) + 3) * sizeof(char));
        strcpy(parts[i]._path, path_prefix);
        strcat(parts[i]._path, num_str);

        free(path_prefix);
    }
}
