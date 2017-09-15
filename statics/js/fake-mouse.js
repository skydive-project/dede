/*
 * Copyright (C) 2017 Red Hat, Inc.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

_DedeFakeMouse = function() {
  var w = window,
      d = document,
      e = d.documentElement,
      g = d.getElementsByTagName('body')[0],
      width = w.innerWidth || e.clientWidth || g.clientWidth,
      height = w.innerHeight|| e.clientHeight|| g.clientHeight;

  this.x = width / 2;
  this.y = height / 2;

  var img = "url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABMAAAAgCAYAAADwvkPPAAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAN1wAADdcBQiibeAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAOBSURBVEiJpdZfSBxHHMDx76yn5lytLWk1D4b6YAu5hMo1lFbSHoRSknuUJrSmpBZJS4ObBPxHCskdxIcgfQpqK01AU6IPCiUNURNNIDGEEGJpY1NoKShNkBqDnt7d3p/1dqcPt17VnD3t/WBYdmf3MzM7uzO/HMABKIAky1CAl/v7+98BnPZ5VvF+OBic6erqOgAUZgt+GQ+FjKVwONbe3n4QKMoG/Cq2uJiwWlulEQzGOjs7a7IBT0cXFkzp8UjrxImsQd8yJl2uFNjR0fHxZkHH2gtidBSHEPlfnDnTA3ymadogoAPWpjEAMTKCA5bBTzVNGwIimcC0WAoUIv9zv7/HMIxPGhoaRjOB62IA4vp1csFZ7/f3AhnB/8RSoBDOep8vI5gRAxDXriV76PP1WpZ1qKmp6UY6cEPYSvCYPeR04IaxFChEwTGfr1dKeai5ufnmSnBTGIAYHiZXiILjp071AavATWMAYmgoOeTVoP6/MBsUeXl5zuMtLX2KonzU2Nh4Y2OYqmLW1xNXVUsWFUkKC6UsKEBRVelU1S3a0aPfV1ZW7lsfU1VQFAiFQNcxdu40f4pGp8bGxu7ruj4XDofDoVAoaBjG35FI5Pd4PP5Heqy8nOi5czLv6lWRc/48AM6LF3NcLS3FHo+nE3gKxAETMIAYEHt+efF4MC5dMgfu3Jk2amokDru9W7d4QVGKu7u7dwHPbPAZsABEAevfXUkIzCNHZOzsWeNka2tHbV1d31/T0zpebxKzLBw9PXn79uypB/JJflumfUztbCdji4sJ49498+nU1Ex1dfVBYBfwoaZpN/VHj0zpcknpckm5e7dMhEJGW1vbBzb4XNTFdD3264MH9ysqKqqArSRX2DccDsfpuSdP4rK2Vi6DiQsXzD8fPhwCXgTESigHyH27qirw3t69/vn5+UkgBCwlR2YVbi0tff1Nr7c0d3hYACiTk6JY08onHz/un5iYmGPNz/6SXbasaakAeKukpOTrSCBgyv37U70zBgeXfhkf/xZQVz6jAIt2ibE6RYgDM7Ozs7/9eOXKlHH4cLJuxw4oK5Pbt22rske24SgC3nW73d8YwaBpXL5sRgOB+A8DA9+VlZW9xjqTsF7kAq8CB8bv3v359sjIiNvt9gLbSeYmqyZApAHW1juBV0i+H53kRxoBEqzJnDJhy/csp12mXdKmX/8Aeoie2lapKEoAAAAASUVORK5CYII=')";

  var img_down = "url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABMAAAAgCAYAAADwvkPPAAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAN1wAADdcBQiibeAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAOMSURBVEiJpdZfSFtXHMDx74nRGDXV0k3CyBgUKwiyEYZjjjEHY7AV9iD2D+tkZD7Y6e5DFaUbtKN9cG1YWMucLg5aW6zWv1DrVmire+i6Eleme6gK20qRrc5Z0pp/mtx5c/aQG9GoS1x+8ONy4J7PPeeewzm/DMAIGABJmmEAnurv738ZMOvttOINf8A/73a79wF56YIf+pf8amA5EG5tbd0PWNIBP1kMLq7Uf1svfSFfuK2t7d10wONPAk80q8sqq/qq0gKNgIg3hmaGAEznHec7pZQORVG+A0JANFVsXQzNDCEQpnMfnLsAbAvcgAEMzgwCxMH3FUW5BiwlAzfF4qAQwtTh6Ligqup7jY2NN5OBW2IAA9MDAOaO2o5uICn4n1gcFAizu9adFEyKAfRP9wOY3bXu7mg0eqipqWl0MzAlbC3YUReb8mZgylgcFELkuOvc3VLKQ83NzWNrwW1hAH1TfQhETntdew+wDtw2BtA71QsCcyL4vzCA3nu9ItuYbT5Tf6bHaDQeaGho+D4lzJJl4WTFSXYYd0QLsgpkfna+zMvKw2KyyFxTbvbhusNdJSUlb22JWbIsGIQBX8RHQA1QZi3T5EP54NbNW+OhUMgbDAaDgUDAr6rqX5FIZNrr9f6+KVa8q5grVVdk11SXOPXjKQBcP7kyzlaczT/22rE24G8gAmiACoSByIbzau+evdxx3NE8NzwPlRcVmWnIBGDk1xGESeR3dnaWAo908BGwCCwD2uqtJBAcfeWovPTOJbXl05avahw1PXN/zIUOlh4EICqjuDyurPLXyz8CTMT2lqY/V2+2jxeDiyujv41q9/+8P19ZWbkfKAWqFEUZm5id0DiB5AQytyVX+pZ9qtPpfFMHN0RNcCkY9kx4xouKisqBXcSO7OeNRuPx2bnZSEVnhYyDzh+c2uTU5DWggDWndDxeGhkZ+Qwo1pEMPa3AgdPO0z8P3xteHZ3tC5sMhoNqdXX1C0BmIrZTz+yEL+UAZYWFhZ8/9j/Wir4sWh3d5V8u/3N34u7XQO7aPgbAp2eY9SVCBJhfWFiYujp89cGRsiMSwG61s3vnbml9xlquzyDlsACv2u32dl/Ip12cuKh5/d7IwODANzabbQ9bLMJWkQk8B+y7PX578vrY9Rt2u/1t4Flitcm6BdiwGgkh9E5PE/s/IWKbdAlYIaFySobF34mXXZqem5Zf/wKFSpLvcZguqQAAAABJRU5ErkJggg==')";

  var sheet = window.document.styleSheets[0];
  sheet.insertRule('#dede-fake-mouse-pointer {position:absolute; z-index:9999999; width:19px; height:32px; background-image: ' + img + ';}', sheet.cssRules.length);
  sheet.insertRule('#dede-fake-mouse-pointer.down {background-image: ' + img_down + ';}', sheet.cssRules.length);

  this.pointer = document.createElement('div');
  this.pointer.setAttribute('id', 'dede-fake-mouse-pointer');
  this.pointer.setAttribute('style', 'left:' + this.x + 'px; top:' + this.y + 'px;');
  document.body.append(this.pointer);
};

_DedeFakeMouse.prototype.easeInOutQuart = function(t) { return t < 0.5 ? 8 * t * t * t * t : 1 - 8 * (--t) * t * t * t; };

_DedeFakeMouse.prototype.click = function(callback) {
	var self = this;

  var addClass = function() {
    self.pointer.className += " dede-mouse-down";
  };

  var removeClass = function() {
    self.pointer.className = self.pointer.className.replace(/(?:^|\s)dede-mouse-down(?!\S)/g, '');
  };

  var i = 4, down = false;
  var blink = function() {
    if (down) {
      removeClass();
      down = false;
    } else {
      addClass();
      down = true;
    }
    if (i-- > 0) {
      setTimeout(blink, 100);
    } else {
    	removeClass();
  		if (callback) callback();
    }
  };

  blink();
};

_DedeFakeMouse.prototype.moveTo = function(x, y, speed, callback) {
  var self = this;

  // move the pointer 1px away to allow selenium to click
  x += 1; y += 1;

  var path = document.createElementNS('http://www.w3.org/2000/svg','path');

  var x1 = this.x, y1 = this.y, x2 = x, y2 = y;
  var xx = x2 - x1, yy = y2 - y1;
  var l = Math.sqrt(xx * xx + yy * yy);
  var cx = (x1 + x2) / 2, cy  = (y1 + y2) / 2;
  var angle = Math.atan2(yy, xx);
  // curve height
  var dist = l * 0.17;
  var cpx = Math.sin(angle) * dist + cx;
  var cpy = -Math.cos(angle) * dist + cy;

  path.setAttribute('d','M ' + x1 + ' ' + y1 + 'Q '+ cpx + ' ' + cpy + ' ' + x2 + ' ' + y2);
  var len = path.getTotalLength();

  var move = function(percent, step) {
    var pt = path.getPointAtLength(percent / 100 * len);
    self.pointer.style.left = pt.x + 'px';
    self.pointer.style.top = pt.y + 'px';

    percent += step;
    if (percent <= 100) {
      var ms = self.easeInOutQuart(percent/100);
			ms = ms * 5 + 5;

      setTimeout(function(){
        move(percent, speed);
      }, ms);
    } else {
      self.x = x;
      self.y = y;

			if (callback) callback();
    }
  };

  move(0, speed);
};

_DedeFakeMouse.prototype.clickTo = function(x, y, callback) {
	var self = this;

	this.moveTo(x, y, 2, function() { self.click(callback); });
};

_DedeFakeMouse.prototype.clickOn = function(el, callback) {
	var self = this;

  this.moveOn(el, function() {
    self.click(callback);
  });
};

_DedeFakeMouse.prototype.moveOn = function(el, callback) {
	var self = this;

  var br = el.getBoundingClientRect();
  this.moveTo((br.left + br.right) / 2, (br.top + br.bottom) / 2, 2, function() {
    // a last move to be as close as possible
    br = el.getBoundingClientRect();
    self.moveTo((br.left + br.right) / 2, (br.top + br.bottom) / 2, 10);

    if (callback) callback();
  });
};

DedeFakeMouse = new _DedeFakeMouse();
