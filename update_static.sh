cd static
rm -fr ./*
git clone --no-checkout https://github.com/pxlsspace/Pxls temp
cd temp
git checkout master -- resources/public
echo ""
echo "Last Pxls commit hash: $(git rev-parse HEAD)"
cd ..
mv -uf temp/resources/public/* .
rm -fr temp
