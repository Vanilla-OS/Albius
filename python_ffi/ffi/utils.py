def __array_repr__(arr, arr_size):
    fmtstr = "[ "
    for i in range(arr_size):
        fmtstr += arr[i].__repr__()
        fmtstr += ", "

    return fmtstr[:-2] + " ]"
