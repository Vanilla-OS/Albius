def __array_repr__(arr):
    fmtstr = "[ "
    for elem in arr:
        fmtstr += elem.__repr__()
        fmtstr += ", "

    return fmtstr[:-2] + " ]"
