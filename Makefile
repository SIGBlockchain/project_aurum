BINS=goTestAll
all: $(BINS)

gotTestAll: gotTestAll.cpp
	g++ -g $? -o $@

clean:
	rm -rf *~ *.dSYM $(BINS)
