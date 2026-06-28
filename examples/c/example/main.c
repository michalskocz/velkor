#include <stdio.h>

int main(void) {
    printf("Hello world !\n");
#ifdef NDEBUG
    printf("Ndebug\n");
#endif

    return 0;
}
