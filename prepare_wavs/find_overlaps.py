# parse flags

import argparse
import json

parser = argparse.ArgumentParser(description='return speech overlap in json to stdout')
parser.add_argument('audio', type=str, help='path audio file')
args = parser.parse_args()

from pyannote.audio import Pipeline
pipeline = Pipeline.from_pretrained("pyannote/overlapped-speech-detection",
                                    use_auth_token="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")
output = pipeline(args.audio)

arr = []

for speech in output.get_timeline().support():
    arr.append({"start": speech.start, "end": speech.end})

final_json = {
    "overlaps": arr
}

print(json.dumps(final_json))
