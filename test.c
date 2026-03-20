int add(int a, int b) {
    return a + b;
}

int main() {
    int x = 5;
    int y = 10;
    int result = add(x, y);

    if (result > 10) {
        result = result - 1;
    } else {
        result = result + 1;
    }

    while (result > 0) {
        result--;
    }

    for (int i = 0; i < 10; i = i + 1) {
        x = x + 1;
    }

    return 0;
}
