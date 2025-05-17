# Notes

Floating Tag Strategy is best for an Actions Repo like this. Basically:

- You create & manage a `v1` branch on the Repo
- Everytime you make changes, you push a new immutable tag up (`v1.0.2` -> `v1.0.3`)
- Afterwards, you force push on the `v1` branch to point it to the commit that the latest tag was set with

So you're still creating immutable tags this way, but you're just updating a v1 branch on the repo to always point at the latest tag that was pushed.

- If individual users want to use new versions, they can do so by just specifying `@v1`
- If they want to use specific tags, they can do so by specifying `@v1.0.2`

## Script

``` sh
#!/bin/bash

# Get the latest v1.x.x tag
latest_tag=$(git tag -l "v1.*" --sort=-creatordate | head -n 1)

if [ -z "$latest_tag" ]; then
  echo "No v1.* tags found"
  exit 1
fi

echo "Updating v1 branch to $latest_tag..."

# Fetch everything just in case
git fetch --all --tags

# Checkout the tag and create a temp branch
git checkout tags/$latest_tag -b tmp-v1

# Force push to v1 branch
git push origin tmp-v1:v1 --force

# Clean up
git checkout main
git branch -D tmp-v1

echo "v1 branch updated to $latest_tag ✅"
```