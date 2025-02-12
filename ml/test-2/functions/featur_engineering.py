import cv2
import numpy as np


def feature_engineer(dataset):
    dataset = calculate_perimeter(dataset)
    dataset = calculate_non_zeros(dataset)
    dataset = calculate_circularity(dataset)
    dataset = calculate_assymetry(dataset)
    dataset = split_channels(dataset)
    
    return dataset

def calculate_perimeter(images_array):
    """It first creates an empty 'perimeter' key in the dictionary. It then iterates over 
    each segmented image in the array and performs the following operations:

    Find the contours in the segmented image using cv2.findContours
    If at least one contour is found, compute the perimeter of the largest contour using cv2.arcLength 
    and assign it to the corresponding 'perimeter' key in the dictionary. If no contours are found, 
    assign 0 to the 'perimeter' key.
    
    Note that the 'perimeter' key is updated with the perimeter of the largest contour in each image. 
    This assumes that the segmented images contain only one object of interest. 
    If multiple objects are present, this function will only calculate the perimeter of the largest object."""
    images_array['perimeter'] = ''
    for i in range(len(images_array['segmented_image'])):
        # Find contours in the binary image
        image = images_array['segmented_image'][i]
        contours, _ = cv2.findContours(image, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        # Compute the perimeter of the largest contour
        images_array['perimeter'][i] = cv2.arcLength(contours[0], True) if len(contours) > 0 else 0
    
    return images_array


def calculate_non_zeros(images_array):
    """It first creates an empty 'non_zeros' key in the dictionary. It then iterates over 
    each segmented image in the array and performs the following operations:

    Get the segmented image from the array
    Count the number of non-zero pixels in the segmented image using np.count_nonzero
    Assign the count to the corresponding 'non_zeros' key in the dictionary
    
    Note that the 'non_zeros' key is updated with the number of non-zero pixels in each segmented image. 
    This can be used as a proxy for the size or area of the segmented object, assuming that the background 
    pixels are set to zero."""
    images_array['non_zeros'] = ''
    for i in range(len(images_array['segmented_image'])):

        image = images_array['segmented_image'][i]
        images_array['non_zeros'][i] = np.count_nonzero(image)

    return images_array


def calculate_circularity(images_array):
    """The circularity is a measure of how closely an object's shape resembles that 
    of a perfect circle. In this function, the circularity of each segmented image 
    is calculated by finding the contour with the maximum area and then computing the 
    ratio of the area of the contour to the square of its perimeter. The result is a 
    scalar value between 0 and 1, where 1 indicates a perfect circle and values 
    approaching 0 indicate increasingly elongated or irregular shapes. 
    
    The circularity values are then stored in a new column named 'circularity' in the 'skin_200' dataframe."""
    images_array['circularity'] = ''
    for i in range(len(images_array['segmented_image'])):
        
        image = images_array['segmented_image'][i]
        
        contours, hierarchy = cv2.findContours(image, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        max_contour = max(contours, key=cv2.contourArea)
        perimeter = cv2.arcLength(max_contour, True)
        area = cv2.contourArea(max_contour)
        images_array['circularity'][i] = (4*np.pi*area) / (perimeter*perimeter)
    
    return images_array


def calculate_assymetry(images_array):
    """Calculates the degree of asymmetry in the given array of segmented images, 
    and stores the main and secondary asymmetry values in new columns named 'main_assymetry' 
    and 'secondary_assymetry' respectively in the 'skin_200' dataframe.

    The degree of asymmetry is calculated using the moments of the contour of each segmented image. 
    The centroid of the contour is first calculated and then the first and second moments are computed 
    relative to the centroid. The degree of asymmetry is then calculated as the ratio of the difference 
    between the moments of the contour along the first and second axes of symmetry, to the sum of those moments.

    The function calculates two separate degrees of asymmetry, one along the first axis of symmetry 
    and another along the second axis of symmetry, and stores the results in two separate columns."""
    images_array['main_assymetry'] = ''
    images_array['secondary_assymetry'] = ''
    for i in range(len(images_array['segmented_image'])):
        
        image = images_array['segmented_image'][i]
        # Find contours in the binary image
        contours, _ = cv2.findContours(image, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        max_contour = max(contours, key=cv2.contourArea)
        
        # Compute the moments of the contour
        moments = cv2.moments(max_contour)

        # Compute the centroid of the contour
        cx = int(moments["m10"] / moments["m00"])
        cy = int(moments["m01"] / moments["m00"])

        # Compute the first and second moments of the contour
        m11 = moments["m11"] - cx * moments["m01"]
        m20 = moments["m20"] - cx * moments["m10"]
        m02 = moments["m02"] - cy * moments["m01"]

        # Compute the degree of asymmetry to the first axis of symmetry
        theta1 = 0.5 * np.arctan2(2 * m11, m20 - m02)
        images_array['main_assymetry'][i] = np.abs(m20 * np.sin(theta1)**2 - 2 * m11 * np.sin(theta1) * np.cos(theta1) + m02 * np.cos(theta1)**2) / (m20 + m02)

        # Compute the degree of asymmetry to the second axis of symmetry
        theta2 = 0.5*np.arctan2(2*m11, -(m20 - m02))
        images_array['secondary_assymetry'][i] = np.abs(m20 * np.sin(theta2)**2 - 2 * m11 * np.sin(theta2) * np.cos(theta2) + m02 * np.cos(theta2)**2) / (m20 + m02)

    return images_array

def split_channels(images_array):
    """Splits the RGB color channels of the images in the given array, crops the images 
    to a specific region of interest, and stores the separated color channels in new columns 
    named 'r_channel', 'g_channel', and 'b_channel' respectively in the 'skin_200' dataframe.

    The function first loops over all the images in the given array and extracts the 
    'image' column from each image. It then crops the image to a specific region of 
    interest using slicing. The cropped image is then split into its three color channels 
    using the cv2.split() function. The separated channels are then stored in separate columns in the 'skin_200' dataframe.

    Note that the function assumes that the input images are in RGB color space."""
    images_array['r_channel'] = ''
    images_array['g_channel'] = ''
    images_array['b_channel'] = ''
    for i in range(len(images_array['image'])):
        
        image = images_array['image'][i]
        image_cropped = image[60:400, 50:550]
        # Use only the pixels that are in the segmented image
        red, green, blue = cv2.split(image_cropped)
        images_array['r_channel'][i] = red
        images_array['g_channel'][i] = green
        images_array['b_channel'][i]  = blue