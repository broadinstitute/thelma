#!/bin/bash

# Files and dirs
RELEASE_TARBALL=${1}
WORKING_DIR=$(basename "${1}" .tar.gz)

# XCode command stuff
APPLE_ID=appledev@broadinstitute.org
TEAM_ID=R787A9V6VV
CMD_AUTH_FLAGS="--apple-id ${APPLE_ID} --password ${APP_PWD} --team-id ${TEAM_ID}"

untar() {
	tar -xf "${1}" -C "${WORKING_DIR}"
}

_tar() {
	tar -czf "${1}" "${2}"
}

sign() {
	codesign -f -o runtime --timestamp -s "5784A30A5BFD511E8636B9F6BBE7EE36D0F0A726" "${1}"
}

_zip() {
	local _outfile="$(basename ${1})".zip
	zip -rq ${_outfile} "${1}"
	echo ${_outfile}
}

notarize() {
	# echo -n "Uploading ${1} for notarization..."
	exec 5>&1
	local _output=$(xcrun notarytool submit ${CMD_AUTH_FLAGS} "${1}" | tee >(cat - >&5))
	# echo "done"

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
			echo "Sleep 15"
			sleep 15
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
		echo "${1} was NOT successfully notarized :("
	fi
}

# Make working dir
mkdir -p "${WORKING_DIR}"

# Untar input
untar "${RELEASE_TARBALL}"

# Sign each binary
echo "Signing binaries..."
for bin in "${WORKING_DIR}"/bin/*
do
	sign "${bin}"
done

# Notarize all files in the working dir
notarize $(_zip "${WORKING_DIR}")

# Verify all files were notarized
for bin in "${WORKING_DIR}"/bin/*
do
	verify "${bin}"
done

# Create SaN release tarball
_tar "san-${WORKING_DIR}.tar.gz" "${WORKING_DIR}"

# Remove working dir and zip file
rm -rf "${WORKING_DIR}" "${WORKING_DIR}.zip"
