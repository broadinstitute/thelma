#!/bin/bash

set -eo pipefail

# This script signs and notarizes release binaries as follows:
# * unpack the given tarball
# * create a keychain and import the signing cert
# * sign each file in the ${RELEASE_DIR}/bin/ directory
# * zip up the whole provided directory (.tar.gz is not supported by Apple)
# * submit the zip to Apple for notarizing
# * verify that all files were notarized
# * create the release tarball

if [[ $# -ne 3 ]]; then
  echo "Usage: $0 <release dir> <release tarball> <output tarball>" >&2
  exit 1
fi

# Check for required env vars
if [ -z "${THELMA_MACOS_APP_PWD}" ]; then
	echo "ERROR: Apple Developer application password env var THELMA_MACOS_APP_PWD unset but required. Exiting."
	exit 1
fi
if [ -z "${THELMA_MACOS_CERT}" ]; then
	echo "ERROR: Signing cert env var THELMA_MACOS_CERT unset but required. Exiting."
	exit 1
fi
if [ -z "${THELMA_MACOS_CERT_PWD}" ]; then
	echo "ERROR: Signing cert password env var THELMA_MACOS_CERT_PWD unset but required. Exiting."
	exit 1
fi

readlinkf(){
	perl -MCwd -e 'print Cwd::abs_path shift' "$1"
}

# Files and dirs
TEMP_DIR=${1}
RELEASE_TARBALL=${2}
OUTPUT_TARBALL=${3}
WORKING_DIR=$(readlinkf "${TEMP_DIR}")/san

# XCode signing info - doesn't contain secrets
APPLE_ID=appledev@broadinstitute.org
TEAM_ID=R787A9V6VV
SECURITY_ID=5784A30A5BFD511E8636B9F6BBE7EE36D0F0A726
CMD_AUTH_FLAGS="--apple-id ${APPLE_ID} --password ${THELMA_MACOS_APP_PWD} --team-id ${TEAM_ID}"

# Sign one file
sign() {
  echo "Signing ${1}..."
	codesign -f -o runtime,library --timestamp -s "${SECURITY_ID}" "${1}"
}

# Zip the given directory into the working dir
archive() {
	# Get the absolute path to the input path
	local _absdir="$(readlinkf "${1}")"

	# Extract just the name of the directory to zip
	local _bname="$(basename ${_absdir})"

	# Save the output filepath
	local _outfile="${WORKING_DIR}/${_bname}".zip

	# Zip will update in-place so make sure to delete 
	rm "${_outfile}" 2>/dev/null

	# Go to the parent directory of the input path to avoid directories inside the zip
	cd "${_absdir}"

	# Create the archive
	zip -rDq "${_outfile}" *

	# Go back to the previous directory
	cd - > /dev/null

	# Other fns get the signed zip file from stdout
	echo "${_outfile}"
}

# Upload the zip file for notarization and wait for a response
notarize() {
	echo "Notarizing ${1}, uploading to Apple..."
	exec 5>&1
	local _output=$(xcrun notarytool submit ${CMD_AUTH_FLAGS} "${1}" | tee >(cat - >&5))

	local _sub_id_regex="id:\ ([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})"
	local _sub_id=""
	if [[ $_output =~ $_sub_id_regex ]]; then
		_sub_id=${BASH_REMATCH[1]}
	else
		echo "${_output}"
		echo "ERROR: A problem occurred while submitting the notarization request. Exiting."
		return 1
	fi

	echo "Checking notarization status for ${_sub_id}"

	local _resp=""
	local _cont=0
	local _wait_total=0
	local _sleep_inc=15
	while
		# Query Apple to see the status of the notarization request
		_resp=$(xcrun notarytool log -f json ${CMD_AUTH_FLAGS} ${_sub_id} 2>&1)
		
		# Most likely response is pending request
		# This message looks like:
		# {
		#	"id": "<UUID>",
		#	"message": "Submission log is not yet available or submissionId does not exist"
		# }
		echo "Response from Apple (${_wait_total}s)"
		echo "-------------------------------------------------"
		echo "${_resp}"
    echo "-------------------------------------------------"

		if echo "${_resp}" | grep -q 'not yet available\|does not exist'; then
			_wait_total=$((_wait_total + _sleep_inc))
			sleep ${_sleep_inc}
		# Eventually the job should complete
		# This message looks like:
		# {
		#   "logFormatVersion": 1,
		#   "jobId": "<submission UUID>",
		#   "status": "Accepted",
		#   "statusSummary": "Ready for distribution",
		#   "statusCode": 0,
		#   "archiveFilename": "<uploaded filename>",
		#   "uploadDate": "<upload date>",
		#   "sha256": "<upload sha256>",
		#   "ticketContents": [
		#     {
		#       "path": "<uploaded filename>/<path to file>",
		#       "digestAlgorithm": "SHA-256",
		#       "cdhash": "<file hash>",
		#       "arch": "<file arch>"
		#     },
		#	  ... (each file has one entry)
		#   ],
		#   "issues": null
		# }
		elif echo "${_resp}" | grep -q 'logFormatVersion'; then
			_cont=1
		else
			echo ${_resp}
			echo "An error occurred during notarization"
		fi
		[[ ${_cont} -eq 0 ]]
	do true; done

	# Add a newline after counting is done
	echo ""

	# Check status field
	if [[ $(echo $_resp | jq -r '.status') == Accepted ]]; then
		echo "Notarization of submission ${_sub_id} complete"
	else
		echo "Notarization of submission ${_sub_id} failed"
		return 1
	fi
}

# Check the notarization status of the given file
verify() {
  echo "Verifying ${1}..."
	_not_info=$(codesign -vvvv -R="notarized" --check-notarization "${1}" 2>&1)
	if echo "${_not_info}" | tr '\n' ' ' | grep -Eq 'valid on disk.*satisfies its Designated Requirement.*explicit requirement satisfied'; then
		echo "${1} was successfully notarized!"
	else
		echo "${_not_info}"
		echo "${1} was NOT successfully notarized :("
	fi
}

# Make working dir
echo "Make working dir ${WORKING_DIR}"
mkdir -p "${WORKING_DIR}"

# Make untar dir
untar_dir="${WORKING_DIR}"/release-files
echo "Unpacking input tarball to ${untar_dir}"
mkdir -p "${untar_dir}"
tar -xf "${RELEASE_TARBALL}" -C "${untar_dir}"

# Sign each binary
echo -n "Signing binaries..."
for bin in "${untar_dir}"/bin/* "${untar_dir}"/tools/bin/*
do
	sign "${bin}"
done
echo "done"

# Submit the release to Apple for notarization
notarize "$(archive "${untar_dir}")"

# Verify all files were notarized
# Note: Oddly, there's no need to check the notarization status of the files
#       that were archived and sent to Apple, as it seems like gatekeeper
#       is able to check files off a specific hash (or similar) rather than
#       relying on some cryptographic information appended to the binaries
#       included in the original zip, which makes sense since we never
#       download anything from Apple post-notarization, and it seems like
#       the temp zip's hash doesn't change before/after notarization, implying
#       that Apple registers the binaries as "safe" and then lets users computers
#       do a check later when they're run....or something. This is all very opaque
#       and the process is unclear.
for bin in "${untar_dir}"/bin/* "${untar_dir}"/tools/bin/*
do
	verify "${bin}"
done

# Create SaN release tarball
echo -n "Creating release tarball ${OUTPUT_TARBALL}..."
tar -C "${untar_dir}" -czf "${OUTPUT_TARBALL}" .
echo "done"

# Remove working dir
rm -rf "${WORKING_DIR}"
