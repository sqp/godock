/*
Package dist contains documentation and files to build distro packages.

Linux package installation.


Debian - Ubuntu - LinuxMint and related from the repository

You can find packages links and install configuration on our repository:

  https://software.opensuse.org//download.html?project=home%3ASQP%3Acairo-dock-go&package=cairo-dock-rework


Create Archlinux package

Install golang applets using the package manager directly from sources.

Create a package with the dock and applets (requires an installed cairo-dock to build).
  mkdir cairo-dock-rework
  cd cairo-dock-rework
  wget https://raw.githubusercontent.com/sqp/godock/master/dist/cairo-dock-rework/PKGBUILD
  makepkg

Or create a package with only applets.
  mkdir cairo-dock-goapplets
  cd cairo-dock-goapplets
  wget https://raw.githubusercontent.com/sqp/godock/master/dist/cairo-dock-goapplets/PKGBUILD
  makepkg

Install command:
  makepkg -i

Remove Package:
  pacman -R cairo-dock-goapplets


Build from sources

Requires go 1.8 or newer (GOPATH from go).

Single applet:

  go get -u github.com/sqp/godock/applets/GoGmail

Applets pack:

  go get -u -d -tags 'gtk all' github.com/sqp/godock/cmd/cdc
  cd $GOPATH/src/github.com/sqp/godock/
  make patch

  make unstable

  # It can then be installed in the system tree.
  # (optional if you add $GOPATH/bin to your PATH)

  sudo make install

Dock rework with applets:

  go get -u -d -tags 'dock all' github.com/sqp/godock/cmd/cdc
  cd $GOPATH/src/github.com/sqp/godock/
  make patch-dock

  make dock

  # or if you want to change applets list:
  make DOCK='dock all' dock

  # It can then be installed in the system tree

  sudo make install-dock

or you can also install manually the applets you need (you may have to restart your dock).

  # Install (make link) for all applets in your home dir.
  for f in $GOPATH/src/github.com/sqp/godock/applets/*; do ln -s $f ~/.config/cairo-dock/third-party/; done

  # Or install just those you need.
  cd $GOPATH/src/github.com/sqp/godock/applets/GoGmail
  make link

  cd $GOPATH/src/github.com/sqp/godock/applets/NetActivity
  make link

The list of applets buildable as standalone can be found in the applets repo:
    https://github.com/sqp/godock/tree/master/applets

The list of applets buildable with the dock or the applets service can be found
in the allapps package:
    https://github.com/sqp/godock/tree/master/services/allapps


Once the dock rework or the applet pack has been installed, the cdc command is
available with a few options:
    http://glx-dock.org/ww_page.php?p=cdc&lang=en


*/
package dist
