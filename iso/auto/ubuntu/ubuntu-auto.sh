#!/usr/bin/env bash
set -o errexit -o pipefail -o noclobber

# define variables
SEED=-
FLAVOR=-
FLAVOR_BIONIC_64=bionic64
FLAVOR_XENIAL_64=xenial64
DOWNLOAD_URL="-"
DOWNLOAD_BIONIC_64_URL="http://releases.ubuntu.com/bionic/ubuntu-18.04.1-live-server-amd64.iso"
DOWNLOAD_XENIAL_64_URL="http://releases.ubuntu.com/xenial/ubuntu-16.04.5-server-amd64.iso"
OLD_ISO_DIR="iso_old"
NEW_ISO_DIR="iso_new"
OLD_ISO_NAME="ubuntu.iso"
NEW_ISO_NAME="ubuntu-auto.iso"
BOOTABLE=n
REUSE=n
WORKSPACE=-
DEBUG=n

# define info function
info()
{
	echo "[INFO] $1"
}

# define debug function
debug()
{
	if [[ $DEBUG == 'y' ]]; then
		echo "[DEBUG] $1"
	fi
}

# define check existence of program
program_is_installed()
{
	local return_=1
    type $1 >/dev/null 2>&1 || { local return_=0; }
    echo $return_
}

# check if current user is root
check_root()
{
	if [ $(id -u) != 0 ]; then
		echo "execute this script as root!"
		exit 1
	fi
}

# define validation
validate()
{
	case "$FLAVOR" in
		"$FLAVOR_BIONIC_64" )
			DOWNLOAD_URL=$DOWNLOAD_BIONIC_64_URL
			;;
		"$FLAVOR_XENIAL_64" )
			DOWNLOAD_URL=$DOWNLOAD_XENIAL_64_URL
			;;
		* )
			echo "value of [-v|--flavor] option is not one of [$FLAVOR_BIONIC_64|$FLAVOR_XENIAL_64]"
			exit 4
	esac
	# SEED
	if [[ $SEED == "-" ]]; then
		echo "[-s|--seed] is not specified"
		exit 4
	elif [[ ! -r $SEED ]]; then
		echo "value of [-s|--seed] option is not readable"
		exit 4
	fi
	# WORKSPACE
	if [[ $WORKSPACE == "-" ]]; then
		echo "[-w|--workspace] is not specified"
		exit 4
	elif [[ ! -d $WORKSPACE ]]; then
		echo "value of [-s|--seed] option is not a directory"
		exit 4
	elif [[ ! -w $WORKSPACE ]]; then
		echo "value of [-s|--seed] option is not a writable directory"
		exit 4
	fi
}

is_mounted()
{
	if grep -qs $1 /proc/mounts; then
		return 1
	else
		return 0
	fi
}

# define clean up function
cleanup()
{
	if grep -qs "$WORKSPACE/$OLD_ISO_DIR" /proc/mounts; then
		debug "unmounting $WORKSPACE/$OLD_ISO_DIR"
		umount "$WORKSPACE/$OLD_ISO_DIR" > /dev/null 2>&1
	fi
	if [[ $REUSE == 'n' ]]; then
		rm -rf "$WORKSPACE/$OLD_ISO_NAME" > /dev/null 2>&1
	fi
	rm -rf "$WORKSPACE/$OLD_ISO_DIR" > /dev/null 2>&1
	rm -rf "$WORKSPACE/$NEW_ISO_DIR" > /dev/null 2>&1

	debug "cleaned up."
}

# define obtain ISO function
obtain_iso()
{
	if [[ $REUSE == 'y' ]] && [[ -e $WORKSPACE/$OLD_ISO_NAME ]]; then
		debug "reused $WORKSPACE/$OLD_ISO_NAME"
		return
	fi
	
	echo "downloading $DOWNLOAD_URL to $WORKSPACE/$OLD_ISO_NAME"
	wget -O $WORKSPACE/$OLD_ISO_NAME $DOWNLOAD_URL
}

# ======================================================
# Parse flags
# Thanks to https://stackoverflow.com/a/29754866/2012567
# ======================================================
# check if getopts can be executed.
! getopt --test > /dev/null 
if [[ ${PIPESTATUS[0]} -ne 4 ]]; then
    echo "`getopt --test` failed in this environment."
    exit 1
fi
# define options
OPTS=s:v:w:drb
LONG_OPTS=seed:,flavor:,workspace:,debug,reuse,bootable
# parse options
! PARSED=$(getopt --options=$OPTS --longoptions=$LONG_OPTS --name "$0" -- "$@")
if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
	echo "`getopt` failed to parse arguments correctly."
	exit 2
fi
eval set -- "$PARSED"
# assign variables
while true; do
	case "$1" in
		-s|--seed)
			SEED="$2"
			shift 2
			;;
		-v|--flavor)
			FLAVOR="$2"
			shift 2
			;;
		-w|--workspace)
			WORKSPACE="$2"
			shift 2
			;;
		-r|--reuse)
			REUSE="y"
			shift
			;;
		-b|--bootable)
			BOOTABLE="y"
			shift
			;;
		-d|--debug)
			DEBUG="y"
			shift
			;;
		--)
			shift
			break
			;;
		*)
			echo "unrecognized argument $1"
			exit 3
			;;
	esac
done

# =======
# Execute
# =======
# validation
validate
debug "options parsed and validated"
debug "SEED=$SEED, FLAVOR=$FLAVOR, DOWNLOAD_URL=$DOWNLOAD_URL, WORKSPACE=$WORKSPACE, REUSE=$REUSE, BOOTABLE=$BOOTABLE"

# preparation
check_root
cleanup
obtain_iso
info 'installation media obtained'

# check dependencies
info "checking required packages"
if [ $(program_is_installed "mkpasswd") -eq 0 ] || [ $(program_is_installed "mkisofs") -eq 0 ]; then
    info "install 'whois' and 'genisoimage'"
    (apt-get -y update > /dev/null 2>&1) &
    (apt-get -y install whois genisoimage > /dev/null 2>&1)
fi
if [[ $BOOTABLE == 'y' ]]; then
	if [ $(program_is_installed "isohybrid") -eq 0 ]; then
		info "install 'syslinux' and 'syslinux-utils'"
		(apt-get -y update > /dev/null 2>&1) &
		(apt-get -y install syslinux syslinux-utils > /dev/null 2>&1)
	fi
fi

info "remastering installation media"
mkdir -p $WORKSPACE/$OLD_ISO_DIR
mkdir -p $WORKSPACE/$NEW_ISO_DIR

# mount installation media
if grep -qs "$WORKSPACE/$OLD_ISO_DIR" /proc/mounts; then
    debug "$WORKSPACE/$OLD_ISO_DIR is already mounted, continue"
else
    (mount -o loop $WORKSPACE/$OLD_ISO_NAME $WORKSPACE/$OLD_ISO_DIR > /dev/null 2>&1)
    debug "$WORKSPACE/$OLD_ISO_DIR mounted"
fi

# copy to new
cp -rT $WORKSPACE/$OLD_ISO_DIR $WORKSPACE/$NEW_ISO_DIR > /dev/null 2>&1
debug "copied from $WORKSPACE/$OLD_ISO_DIR to $WORKSPACE/$NEW_ISO_DIR"

# set language
echo en > $WORKSPACE/$NEW_ISO_DIR/isolinux/lang
debug "language set to en"

# update timeout
sed -i -r 's/timeout\s+[0-9]+/timeout 1/g' $WORKSPACE/$NEW_ISO_DIR/isolinux/isolinux.cfg
debug "menu timeout updated to 1"

# copy seed file
cp -rT $SEED $WORKSPACE/$NEW_ISO_DIR/preseed/imulab.seed
debug "seed file $SEED copied to $WORKSPACE/$NEW_ISO_DIR/preseed/imulab.seed"

# calculate checksum
seed_checksum=$(md5sum $WORKSPACE/$NEW_ISO_DIR/preseed/imulab.seed | awk '{ print $1 }')
debug "seed file $SEED checksum is $seed_checksum"

# update menu
sed -i "/label live/ilabel autoinstall\n\
  menu label ^Autoinstall Imulab Ubuntu Server\n\
  kernel /install/vmlinuz\n\
  append file=/cdrom/preseed/ubuntu-server.seed initrd=/install/initrd.gz auto=true priority=high preseed/file=/cdrom/preseed/imulab.seed preseed/file/checksum=$seed_checksum --" $WORKSPACE/$NEW_ISO_DIR/isolinux/txt.cfg
debug "isolinux cfg file $WORKSPACE/$NEW_ISO_DIR/isolinux/txt.cfg updated"

# create iso image
debug "making new iso image"
pushd $WORKSPACE/$NEW_ISO_DIR > /dev/null
	mkisofs -D -r -V "IMULAB_UBUNTU" \
		 -cache-inodes -J -l \
		 -b isolinux/isolinux.bin \
		 -c isolinux/boot.cat \
		 -no-emul-boot \
		 -boot-load-size 4 \
		 -boot-info-table \
		 -o $WORKSPACE/$NEW_ISO_NAME . > /dev/null 2>&1
popd > /dev/null

# clean before exit
if [[ $DEBUG == 'n' ]]; then
	cleanup
fi

info "SUCCESS: auto installation media remastered at $WORKSPACE/$NEW_ISO_NAME"