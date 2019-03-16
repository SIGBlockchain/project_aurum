#include <cstring>
#include <iostream>
#include <cstdlib>

using namespace std;

int main() {
    FILE *f;
    if(getenv("windir"))
        f = popen("go test -v .\\...", "r");
    else
        f = popen("go test -v ./...", "r");
        //f = popen("make clean", "r");

    return 0;
}
