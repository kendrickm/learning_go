#version 330 core
out vec4 FragColor;

in vec2 TexCoord;
in vec3 Normal;
in vec3 FragPos;


uniform sampler2D texture1;
uniform vec3 viewPos;
uniform vec3 lightPos;
uniform vec3 lightColor; //Color and brightness
uniform vec3 ambientLight;

void main(){

    vec3 ambientLight = lightColor * 0.3;

    // diffuse
    vec3 lightDir = normalize(lightPos - FragPos);
    float diff = max(dot(Normal,lightDir),0.0);
    vec3 diffuse = diff * lightColor;

    //specular
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir,Normal);
    float spec = pow(max(dot(viewDir,reflectDir), 0.0),32);
    vec3 specular = 0.2 * spec * lightColor;



	FragColor = vec4(ambientLight+diffuse+specular,1.0) * texture(texture1,TexCoord);
}