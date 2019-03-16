BINS=goTestAll
all: $(BINS)

goTestAll: goTestAll.cpp
	g++ -g $? -o $@

run: ${BINS}
	./${BINS}

clean:
	rm -rf *~ *.dSYM $(BINS)
