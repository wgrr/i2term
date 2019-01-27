/*

lukeidraw takes an image file (it supports jpeg, png, gif, bmp, tiff,
vp8l and webp encodings) and prints a formated output that has image
dimensions and current terminal position in terminal character size.
The output is meant to be parsed by external programs, its output is
formated as following:
	imageWidth imageHeight termCurrentRow termCurrentColumn

Example:
	lukeidraw img.png # 30 96 39 1

*/
package main
