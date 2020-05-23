#! /bin/bash

tiltfiletemplate="$(pwd)/Tiltfile.template"
tiltfile="$(pwd)/Tiltfile"
if [ ! -f "$tiltfile" ]; then
	printf '%s\n' "allow_k8s_contexts('sysflow/api-marshall-os-fyre-ibm-com:6443/maryang')" | cat - $tiltfiletemplate > $tiltfile
fi

tilt up --no-browser
