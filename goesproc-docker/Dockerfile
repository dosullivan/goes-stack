FROM ubuntu:22.04

# Set environment variables to avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# Install dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    libopencv-dev \
    zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

# Clone and build goestools
WORKDIR /usr/src
RUN git clone --recursive https://github.com/pietern/goestools.git
WORKDIR /usr/src/goestools
RUN mkdir -p build
WORKDIR /usr/src/goestools/build
RUN cmake ../ -DCMAKE_INSTALL_PREFIX=/usr/local
RUN make -j$(nproc)
RUN make install

# Create working directory for data
WORKDIR /data

# Set default command
CMD ["goesproc", "--help"]