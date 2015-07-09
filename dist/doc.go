/*
Package dist contains documentation and files to build distro packages.

Linux package installation.


Install Debian or Ubuntu package from repository

/etc/apt/sources.list.d/cairo-dock-go.list
  deb http://download.opensuse.org/repositories/home:/SQP:/cairo-dock-go/Debian_8.0/ ./


Install Archlinux package from repository

/etc/pacman.conf
  [home_SQP_cairo-dock-go_Arch_Extra]
  SigLevel = Never
  Server = http://download.opensuse.org/repositories/home:/SQP:/cairo-dock-go/Arch_Extra/$arch


Build Archlinux package

Install golang applets using the package manager directly from sources.

Create a package with the dock and applets (requires an installed cairo-dock to build).
  mkdir cairo-dock-rework
  cd cairo-dock-rework
  wget https://raw.githubusercontent.com/sqp/godock/master/dist/archlinux/cairo-dock-rework/PKGBUILD
  makepkg

Or create a package with only applets.
  mkdir cairo-dock-goapplets
  cd cairo-dock-goapplets
  wget https://raw.githubusercontent.com/sqp/godock/master/dist/archlinux/cairo-dock-goapplets/PKGBUILD
  makepkg

Install package
  makepkg -i

Remove Package
  pacman -R cairo-dock-control


Install from sources

Single applet:

  go get -u github.com/sqp/godock/applets/GoGmail

Applets pack:

  go get -u -d -tags 'gtk all' github.com/sqp/godock/cmd/cdc
  cd $GOPATH/src/github.com/sqp/godock/
  make patch
    make unstable

    # It can then be installed in the system tree (optional)

    make install

Dock rework with applets:

  go get -u -d -tags 'dock all' github.com/sqp/godock/cmd/cdc
  cd $GOPATH/src/github.com/sqp/godock/
  make patch
    make dock

    # or if you want to change applets list:
    make DOCK='dock all' dock

    # It can then be installed in the system tree (optional)

    make install-dock

or you can also install manually the applets you need (you may have to restart your dock).

  cd $GOPATH/src/github.com/sqp/godock/applets/GoGmail
  make link

  cd $GOPATH/src/github.com/sqp/godock/applets/NetActivity
  make link

The list of buildable applets can be found in the applets repo:
    https://github.com/sqp/godock/tree/master/applets

Once the dock rework or the applet pack has been installed, the cdc command is
available with a few options:
    http://glx-dock.org/ww_page.php?p=cdc&lang=en


*/
package dist
