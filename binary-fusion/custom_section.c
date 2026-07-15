#include <stdio.h>

__attribute__((section(".my_custom_section")))
char customData[] = "Hello";

int main() {
    printf("Custom section test\n");
    return 0;
}