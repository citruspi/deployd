## deployd

deployd is a program I (plan to) use for deploying projects to different servers. It currently supports static websites and will support backend services written in Python and Go in the future.

deployd is opinionated - I will be using it for my projects on my servers, so it's designed to work with them.

## Installation

### Manual

Dependencies for deployd are managed with pote's [gpm](https://github.com/pote/gpm).

    $ git clone https://github.com/citruspi/deployd.git
    $ cd deployd
    $ gpm
    $ go build deployd.git
    $ mv deployd /usr/local/bin/.

### Pre-Built Binary (Linux 64-bit)

The continuous integration server (Jenkins running on 64-bit CentOS 7) builds a binary for each commit to the master branch. The binaries are zipped up and uploaded to S3.

    $ wget https://s3.amazonaws.com/deployd/master-latest.zip
    $ unzip master-latest.zip
    $ mv deployd /usr/local/bin/.   
