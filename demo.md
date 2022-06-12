
# Service descriptor

## Publish to n number
curl -X POST -H "Content-Type: application/json" \
    -d '{"Message": "Hey all", "To": ["XXXXXXXX" "XXXXXXXX"]}' \
    localhost:8010/messagesPublish

## Publish n messages

curl -X POST -H "Content-Type: application/json" \
    -d '[{"Message": "First msg", "To": "XXXXXXXX"}, {"Message": "Second msg", "To": "XXXXXXXX"}]' \
    localhost:8010/messagesBulk   

## Publish one msg

curl -X POST -H "Content-Type: application/json" \
    -d '{"Message": "First msg", "To": "XXXXXXXX"}' \
    localhost:8010/messages 

