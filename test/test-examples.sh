#!/usr/bin/env bash
# check that our examples still build, even if we don't run them here

# shellcheck disable=SC1091
. test/util.sh

echo running "$0"

ROOT=$(dirname "${BASH_SOURCE}")/..
cd "${ROOT}"

failures=''

# Test examples/lang/ directory to see if the .mcl files compile correctly.

find_mcl_examples() {
	git ls-files | grep '\.mcl$' | grep '^examples/lang/' | grep -v 'modules/'
}

for file in $(find_mcl_examples); do
	#echo "mcl: $file"
	run-test ./mgmt run --tmp-prefix lang --only-unify "$file" &> /dev/null || fail_test "could not compile: $file"
done

buildout='test-examples.out'
# make symlink to outside of package
linkto="`pwd`/examples/lib/"
tmpdir="`$mktemp --tmpdir -d tmp.XXX`"	# get a dir outside of the main package
cd "$tmpdir"
ln -s "$linkto"	# symlink outside of dir
cd `basename "$linkto"`

# loop through individual *.go files in working dir
for file in `find . -maxdepth 9 -type f -name '*.go'`; do
	#echo "running test on: $file"
	run-test go build -o "$buildout" "$file" || fail_test "could not build: $file"
done
rm "$buildout" || true	# clean up build mess

cd - >/dev/null	# back to tmp dir
rm `basename "$linkto"`
cd ..
rmdir "$tmpdir"	# cleanup

if [[ -n "$failures" ]]; then
	echo 'FAIL'
	echo "The following tests (in: ${linkto}) have failed:"
	echo -e "$failures"
	echo
	exit 1
fi
echo 'PASS'
