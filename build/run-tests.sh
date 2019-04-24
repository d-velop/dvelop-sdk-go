#!/bin/bash

# cf. https://stackoverflow.com/questions/1489277/how-to-use-prune-option-of-find-in-sh for find with prune
# cf. https://stackoverflow.com/a/4667725 for process substitution < < (find...)
# cf. https://dave.cheney.net/2018/07/16/using-go-modules-with-travis-ci for travis an go modules

while read f
do
    cd ${f}; GO111MODULE=on go test ./... ; (( exit_status = exit_status || $? ))
done < <(find $PWD -maxdepth 1 \( -name .git -o -name .idea -o -name build \) -prune -o ! -path $PWD -type d -print )

exit ${exit_status}
