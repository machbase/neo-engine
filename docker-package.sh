# Build package to deploy for aws ec2 instance (amazone linux 2)
#       it is demanded beacuase of cgo cross-compilation.
#       gorm drivers requires native libraries.
#
# 0. Base Amazon linux2 image
#
# 1. Install required packages
#    - golang
#    - which
#    - zip
#    - make
#
# 2. Add additional nameserver to /etc/resolv.conf
#
# 3. set .netrc and GOPRIVATE to access private repo of github.com/machbase
#

PKGNAME=$1
OS=$2
ARCH=$3

# PKGNAME="machgo"
# OS="linux"
# ARCH="arm64"

IMAGE="${PKGNAME}_buildenv_${OS}_${ARCH}:latest"

docker image inspect $IMAGE --format "Check $IMAGE exists." 2> /dev/null
exists=$?
if [ $exists -ne 0 ]; then
    echo "Creating docker image for build environment ..."
    docker build -f scripts/buildenv-dockerfile -t $IMAGE  --platform $OS/$ARCH .
fi

if [ ! -f ~/.netrc ]; then
    echo "~/.netrc not found, it is required to check-out dependency modules"
    exit 1
fi

echo "Build package via $IMAGE"
docker run \
    --rm \
    --platform $OS/$ARCH \
    -v "$PWD":/machgo \
    -w /machgo \
    -v "$HOME/.netrc":/root/.netrc \
    -v "$HOME/go:/root/go" \
    -e GOPRIVATE="github.com/machbase/*" \
    $IMAGE \
    /bin/bash -c "./scripts/package.sh $PKGNAME $OS $ARCH"
