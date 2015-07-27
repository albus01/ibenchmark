TARGET=iBench

all: $(TARGET)

iBench: iBench.go
	go build iBench.go


clean:
	rm -f $(TARGET)

distclean: clean
	rm -rf .ibenchmark
