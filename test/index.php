<?php
require_once __DIR__.'/vendor/autoload.php'; 
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Exception;
$app = new Silex\Application(); 

$app->get('/hello/{name}', function($name) use($app) { 
      return 'Hello '.$app->escape($name); 
}); 

$app->get('/301', function() use($app) {
  header('Location: /hello/redirect ');  
  throw new Exception\HttpException(301, '');
}); 


$app->run();
