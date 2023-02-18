all:
	go build -o albius.so -buildmode=c-shared

clean:
	rm albius.so
