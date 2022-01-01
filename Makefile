build: 
	go build -o ./tree ./tree.go

run: build
	./tree
