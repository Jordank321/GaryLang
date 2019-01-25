cmp [$varName], byte 0
je allocate

mov rcx, [$varName]
call free

allocate:

mov rcx, $valLength
call malloc
mov [$varName], rax
mov rcx, [$value]
mov [rax], rcx
