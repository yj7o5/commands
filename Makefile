build: 
	go build -o ./out/tree ./tree/tree.go

run: build
	./out/tree
