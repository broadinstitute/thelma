#!/bin/bash

# This script signs and notarizes release binaries as follows:
# * sign each file in the ${RELEASE_DIR}/bin/ directory
# * zip up the whole provided directory (.tar.gz is not supported by Apple)
# * submit the zip to Apple for notarizing
# * verify that all files were notarized
# * create the release tarball

# Files and dirs
RELEASE_DIR=${1}
RELEASE_TARBALL=${2}
WORKING_DIR=$(dirname "$(readlink -f "${RELEASE_DIR}")")/san

# XCode command stuff
APPLE_ID=appledev@broadinstitute.org
TEAM_ID=R787A9V6VV
CMD_AUTH_FLAGS="--apple-id ${APPLE_ID} --password ${APP_PWD} --team-id ${TEAM_ID}"

_tar() {
	tar -C "${1}" -czf "${2}" .
}

sign() {
	codesign -f -o runtime --timestamp -s "5784A30A5BFD511E8636B9F6BBE7EE36D0F0A726" "${1}"
}

_zip() {
	# Get the absolute path to the input path
	local _absdir="$(readlink -f "${1}")"

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

	echo -n "Checking notarization status for ${_sub_id}"

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
		if echo "${_resp}" | grep -q 'not yet available\|does not exist'; then
			if [[ ${_wait_total} > 0 ]]; then
				echo -n "...${_wait_total}"
			fi
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

verify() {
	_not_info=$(codesign -vvvv -R="notarized" --check-notarization "${1}" 2>&1)
	if echo "${_not_info}" | tr '\n' ' ' | grep -Eq 'valid on disk.*satisfies its Designated Requirement.*explicit requirement satisfied'; then
		echo "${1} was successfully notarized!"
	else
		echo "${_not_info}"
		echo "${1} was NOT successfully notarized :("
	fi
}

# Make working dir
echo "Making ${WORKING_DIR}"
mkdir -p "${WORKING_DIR}"

# Sign each binary
echo "Signing binaries..."
for bin in "${RELEASE_DIR}"/bin/*
do
	sign "${bin}"
done

# Submit the release to Apple for notarization
notarize "$(_zip "${RELEASE_DIR}")"

# Verify all files were notarized
for bin in "${RELEASE_DIR}"/bin/*
do
	verify "${bin}"
done

# Create SaN release tarball
echo -n "Creating release tarball ${RELEASE_TARBALL}..."
tar -C "${RELEASE_DIR}" -czf "${RELEASE_TARBALL}" .
echo "done"

# Remove working dir
rm -rf "${WORKING_DIR}"
