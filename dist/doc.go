/*
Linux package installation.


Archlinux package

Install golang applets using the package manager directly from sources.

Create a package with the dock and applets (requires an installed cairo-dock to build).
  mkdir cdc-test
  cd cdc-test
  wget https://raw.githubusercontent.com/sqp/godock/master/dist/archlinux/dock/PKGBUILD
  makepkg

Or create a package with only applets.
  mkdir cdc-test
  cd cdc-test
  wget https://raw.githubusercontent.com/sqp/godock/master/dist/archlinux/PKGBUILD
  makepkg

Install package
  makepkg -i

Remove Package
  pacman -R cairo-dock-control
*/
package dist
