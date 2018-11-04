<?php
$lines = file('output.txt');

$labels = array();
$gdax = array();
$bitfinex = array();
$bitmex = array();
array_shift($lines);

$cutoff = $argv[1];
if (empty($cutoff)) {
  $cutoff = -30;
} else {
  $cutoff = 0 - $cutoff;
}

$length = $argv[2];
if (empty($length)) {
  $lines = array_slice($lines, $cutoff);
} else {
  $lines = array_slice($lines, $cutoff, $length);
}

foreach ($lines as $line) {
  $arrLine = explode('###', $line);
  $labels[] = $arrLine[0];
  $gdax[] = array($arrLine[1], $arrLine[2]);
  $bitfinex[] = array($arrLine[3], $arrLine[4]);
  $bitmex[] = array($arrLine[5], $arrLine[6]);
}
unset($lines);
$strOut = "var labels = [";
foreach ($labels as $l) {
  $strOut .= "'$l',";
}
$strOut .= "];\n\n";

$strOut .= "var datasets = [";
// Gdax
$strOut .= "{'label': 'Gdax Buy', 'borderColor': '#009900', 'fill': false, 'data' :[";
foreach ($gdax as $l) {
  $strOut .= "{$l[0]},";
}
$strOut .= "]},\n";
$strOut .= "{'label': 'Gdax Sell', 'borderColor': '#990000', 'fill': '-1', 'data' :[";
foreach ($gdax as $l) {
  $strOut .= "{$l[1]},";
}
$strOut .= "]},\n";
// Bitfinex
$strOut .= "{'label': 'Bitfinex Buy', 'borderColor': '#00CC00', 'fill': false, 'data' :[";
foreach ($bitfinex as $l) {
  $strOut .= "{$l[0]},";
}
$strOut .= "]},\n";
$strOut .= "{'label': 'Bitfinex Sell', 'borderColor': '#CC0000', 'fill': '-1', 'data' :[";
foreach ($bitfinex as $l) {
  $strOut .= "{$l[1]},";
}
$strOut .= "]},\n";
// Bitmex
$strOut .= "{'label': 'Bitmex Buy', 'borderColor': '#00FF00', 'fill': false, 'data' :[";
foreach ($bitmex as $l) {
  $strOut .= "{$l[0]},";
}
$strOut .= "]},\n";
$strOut .= "{'label': 'Bitmex Sell', 'borderColor': '#FF0000', 'fill': '-1', 'data' :[";
foreach ($bitmex as $l) {
  $strOut .= "{$l[1]},";
}
$strOut .= "]},\n";

$strOut .= "];\n\n";

file_put_contents('dataset.js', $strOut);
