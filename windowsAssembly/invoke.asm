; ---------------------------------------------------------------------------
; Define macro: Invoke
; ---------------------------------------------------------------------------
%macro Invoke 1-*
        %if %0 > 1
                %rotate 1
                mov rcx,qword %1
                %rotate 1
                %if %0 > 2
                        mov rdx,qword %1
                        %rotate 1
                        %if  %0 > 3
                                mov r8,qword %1
                                %rotate 1
                                %if  %0 > 4
                                        mov r9,qword %1
                                        %rotate 1
                                        %if  %0 > 5
                                                %assign max %0-5
                                                %assign i 32
                                                %rep max
                                                        mov rax,qword %1
                                                        mov qword [rsp+i],rax
                                                        %assign i i+8
                                                        %rotate 1
                                                %endrep
                                        %endif
                                %endif
                        %endif
                %endif
        %endif
        ; ------------------------
        ; call %1 ; would be the same as this:
        ; -----------------------------------------
        sub rsp,qword 8
        mov qword [rsp],%%returnAddress
        jmp %1
        %%returnAddress:
        ; -----------------------------------------
%endmacro