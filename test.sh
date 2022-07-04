trap catch_errors ERR;

function catch_errors() {
   echo "Tests failed";
   exit 1;
}

cat <<EOF > test-data
Hello world
EOF

PASSWORD=qwerty

echo "===== Encrypting files"
simple-gpg -password $PASSWORD test-data
simple-gpg -password $PASSWORD -outputFile yolo.me test-data

echo
echo "===== Dencrypting files"
simple-gpg -decrypt -password $PASSWORD test-data.gpg
simple-gpg -decrypt -password $PASSWORD yolo.me

echo
echo "===== Checking file output"

cat decrypted-test-data | grep "Hello"
cat decrypted-yolo.me | grep "Hello"

echo
echo "===== Cleaning up"
rm -v test-data
rm -v test-data.gpg
rm yolo.me
rm decrypted-test-data
rm decrypted-yolo.me
