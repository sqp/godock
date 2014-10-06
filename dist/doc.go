/*
Linux package installation.


Archlinux package

Install golang applets using the package manager directly from sources.

Create package
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
