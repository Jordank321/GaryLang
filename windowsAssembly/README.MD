# Testing out NASM here...

## helloworld build
- nasm.exe helloworld.asm -f win64 -o helloworld.obj
- gcc.exe helloworld.obj -m64 -o helloworld.exe