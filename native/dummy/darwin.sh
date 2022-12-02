	gcc -I.. -c machengine_dummy.c
    if [ -f ../libmachengine_dummy_darwin_amd64.a ]; then
        rm ../libmachengine_dummy_darwin_amd64.a
    fi
	ar rc ../libmachengine_dummy_darwin_amd64.a machengine_dummy.o