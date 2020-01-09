FROM archlinux
WORKDIR /app

RUN pacman -Syu go imagemagick fontconfig base-devel --noconfirm

COPY . .

ENV CGO_CFLAGS_ALLOW '-fopenmp'
ENV LD_LIBRARY_PATH /usr/local/lib

RUN go build \
        -o bin/go-image-processor-httpapi \
        main.go

CMD ["bin/go-image-processor-httpapi"]