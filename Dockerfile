FROM golang:latest

ENV DEBIAN_FRONTEND=noninteractive


# tookchain
RUN apt-get update && apt-get install -y \
    git \
    make \
    cmake \
    gcc \
    g++ \
    golang \
    python3 \
    python3-pip \
	python3.11-venv \
file\
    && rm -rf /var/lib/apt/lists/*


# python & pip
RUN if [ ! -e /usr/bin/python ]; then ln -s /usr/bin/python3 /usr/bin/python; fi && \
    if [ ! -e /usr/bin/pip ]; then ln -s /usr/bin/pip3 /usr/bin/pip; fi

WORKDIR /app/opml-opt

COPY . .

RUN python3 -m venv /app/venv
ENV PATH="/app/venv/bin:$PATH"

RUN echo "Repository contents:" && \
    ls -la && \
    echo "Unicorn directory contents:" && \
    ls -la unicorn

RUN go mod download

RUN cd mlgo && pip install -r requirements.txt

WORKDIR /app/opml-opt

RUN make build

ENV CONFIG_FILE=/app/config.yml
ENV DEPENDENCY_FILE=/app/dependency_file

RUN echo "Build complete. Directory contents:" &&  ls -la

CMD ["sh", "-c", "./opt -config $CONFIG_FILE"]
