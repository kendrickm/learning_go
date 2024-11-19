#version 330 core
out vec4 FragColor;

in vec2 TexCoord;
in vec3 Normal;
in vec3 FragPos;


uniform sampler2D texture1;
uniform vec3 lightPos;
uniform vec3 lightColor; //Color and brightness

void main(){
	vec3 ambientLight = vec3(.3,.3,.3);

	//diffuse light
	vec3 lightDir = normalize(lightPos - FragPos);
	float diff = max(dot(Normal,lightDir),0.0);
	vec3 diffuse = diff * lightColor;

	FragColor = vec4(ambientLight + diffuse,1.0) * texture(texture1,TexCoord);
}