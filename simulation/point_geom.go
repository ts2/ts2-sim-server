/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 TEAM
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package simulation

import ()

/*
Point type represents a point on the scenery
*/
type Point struct {
	X float64
	Y float64
}

func(p Point) Xi() int {
	return int(p.X)
}
func(p Point) Yi() int {
	return int(p.Y)
}
func Add(p1 Point, p2 Point) Point {
	return Point{p1.X + p2.X, p1.Y + p2.Y}
}

