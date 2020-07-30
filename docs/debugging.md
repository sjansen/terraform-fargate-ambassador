aws sqs send-message \
    --queue-url `terraform output queue_url` \
    --message-body 'The cake is a lie.'

curl ${AMBASSADOR}/status

curl -d msg="Kilroy was here." \
    ${APPLICATION}/echo
